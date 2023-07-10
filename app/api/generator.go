package api

import (
	"context"

	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/service"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/telemetry"
	"google.golang.org/grpc"
)

type Generator struct {
	logger telemetry.Logger
	proto.UnimplementedGeneratorServer
	uniqueNumberGeneratorFactory service.UniqueNumberGenFactory
	uniqueNumberGenerators       map[string]*service.UniqueNumberGen
	uniqueStringGenerators       map[string]*service.UniqueStringGen
}

var _ proto.GeneratorServer = (*Generator)(nil)
var _ runner.Service = (*Generator)(nil)

func (g Generator) Start(runner *runner.ServiceRunner) *errs.Error {
	runner.WithGRPCServer(func(server *grpc.Server) {
		proto.RegisterGeneratorServer(server, g)
	})
	return nil
}

func (g Generator) GenerateUniqueNumber(
	ct context.Context,
	request *proto.GenerateUniqueNumberRequest,
) (*proto.GenerateUniqueNumberResponse, error) {
	uniqueNumGen, ok := g.uniqueNumberGenerators[request.SequenceName]
	if !ok {
		var err *errs.Error
		uniqueNumGen, err = g.uniqueNumberGeneratorFactory.MakeUniqueNumberGen(request.SequenceName)
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

	return &proto.GenerateUniqueNumberResponse{UniqueNumber: uniqueNum}, nil
}

func (g Generator) GenerateUniqueString(
	ct context.Context,
	request *proto.GenerateUniqueStringRequest,
) (*proto.GenerateUniqueStringResponse, error) {
	uniqueStringGen, ok := g.uniqueStringGenerators[request.SequenceName]
	if !ok {
		strGen, err := service.NewUniqueStringGen(
			request.SequenceName,
			int(request.StringLength),
			request.Alphabet,
			g.uniqueNumberGeneratorFactory)
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

	return &proto.GenerateUniqueStringResponse{UniqueString: uniqueStr}, nil
}

func NewGenerator(
	logger telemetry.Logger,
	uniqueNumberGeneratorFactory service.UniqueNumberGenFactory,
) Generator {
	return Generator{
		logger:                       logger,
		uniqueNumberGeneratorFactory: uniqueNumberGeneratorFactory,
		uniqueNumberGenerators:       make(map[string]*service.UniqueNumberGen),
		uniqueStringGenerators:       make(map[string]*service.UniqueStringGen),
	}
}
