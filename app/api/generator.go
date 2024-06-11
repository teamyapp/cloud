package api

import (
	"context"

	"github.com/teamyapp/cloud/app/service"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/telemetry"
	pbcloud "github.com/teamyapp/protocol/pb/pbgo/cloud"
	"google.golang.org/grpc"
)

type Generator struct {
	logger telemetry.Logger
	pbcloud.UnimplementedGeneratorServer
	uniqueNumberGeneratorRegistry *service.UniqueNumberGenRegistry
	uniqueNumberGenerators        map[string]*service.UniqueNumberGen
	uniqueStringGenerators        map[string]*service.UniqueStringGen
}

var _ pbcloud.GeneratorServer = (*Generator)(nil)
var _ runner.Service = (*Generator)(nil)

func (g *Generator) Start(runner *runner.ServiceRunner) *errs.Error {
	runner.WithGRPCServer(func(server *grpc.Server) {
		pbcloud.RegisterGeneratorServer(server, g)
	})
	return nil
}

func (g *Generator) GenerateUniqueNumber(
	ct context.Context,
	request *pbcloud.GenerateUniqueNumberRequest,
) (*pbcloud.GenerateUniqueNumberResponse, error) {
	uniqueNumGen, ok := g.uniqueNumberGenerators[request.SequenceName]
	if !ok {
		var err *errs.Error
		uniqueNumGen, err = g.uniqueNumberGeneratorRegistry.GetUniqueNumberGen(request.SequenceName)
		if err != nil {
			g.logger.ErrorWithContext(ct, err)
			return nil, errs.ToGRPCErr(err)
		}

		g.uniqueNumberGenerators[request.SequenceName] = uniqueNumGen
	}

	uniqueNum, err := uniqueNumGen.GenerateUniqueNumber(ct)
	if err != nil {
		g.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &pbcloud.GenerateUniqueNumberResponse{UniqueNumber: uniqueNum}, nil
}

func (g *Generator) GenerateUniqueString(
	ct context.Context,
	request *pbcloud.GenerateUniqueStringRequest,
) (*pbcloud.GenerateUniqueStringResponse, error) {
	uniqueStringGen, ok := g.uniqueStringGenerators[request.SequenceName]
	if !ok {
		strGen, err := service.NewUniqueStringGen(
			request.SequenceName,
			int(request.StringLength),
			request.Alphabet,
			g.uniqueNumberGeneratorRegistry)
		if err != nil {
			g.logger.ErrorWithContext(ct, err)
			return nil, errs.ToGRPCErr(err)
		}
		uniqueStringGen = &strGen
		g.uniqueStringGenerators[request.SequenceName] = uniqueStringGen
	}

	uniqueStr, err := uniqueStringGen.GenerateUniqueString(ct)
	if err != nil {
		g.logger.ErrorWithContext(ct, err)
		return nil, errs.ToGRPCErr(err)
	}

	return &pbcloud.GenerateUniqueStringResponse{UniqueString: uniqueStr}, nil
}

func NewGenerator(
	logger telemetry.Logger,
	uniqueNumberGeneratorRegistry *service.UniqueNumberGenRegistry,
) *Generator {
	return &Generator{
		logger:                        logger,
		uniqueNumberGeneratorRegistry: uniqueNumberGeneratorRegistry,
		uniqueNumberGenerators:        make(map[string]*service.UniqueNumberGen),
		uniqueStringGenerators:        make(map[string]*service.UniqueStringGen),
	}
}
