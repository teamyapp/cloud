package api

import (
	"context"

	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/gen"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/cloud/libs/runner"
	"google.golang.org/grpc"
)

type Generator struct {
	dataCollector obs.DataCollector
	proto.UnimplementedGeneratorServer
	uniqueNumberGeneratorFactory gen.UniqueNumberFactory
	uniqueNumberGenerators       map[string]*gen.UniqueNumber
	uniqueStringGenerators       map[string]*gen.UniqueString
}

var _ proto.GeneratorServer = (*Generator)(nil)
var _ runner.Service = (*Generator)(nil)

func (g Generator) Start(runner *runner.ServiceRunner) error {
	runner.WithGRPCServer(func(server *grpc.Server) {
		proto.RegisterGeneratorServer(server, g)
	})
	return nil
}

func (g Generator) GenerateUniqueNumber(
	ctx context.Context,
	request *proto.GenerateUniqueNumberRequest,
) (*proto.GenerateUniqueNumberResponse, error) {
	uniqueNumGen, ok := g.uniqueNumberGenerators[request.SequenceName]
	if !ok {
		var err error
		uniqueNumGen, err = g.uniqueNumberGeneratorFactory.MakeUniqueNumber(request.SequenceName)
		if err != nil {
			g.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
			return nil, err
		}

		g.uniqueNumberGenerators[request.SequenceName] = uniqueNumGen
	}

	uniqueNum, err := uniqueNumGen.GenerateUniqueNumber()
	if err != nil {
		g.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	return &proto.GenerateUniqueNumberResponse{UniqueNumber: uniqueNum}, nil
}

func (g Generator) GenerateUniqueString(
	ctx context.Context,
	request *proto.GenerateUniqueStringRequest,
) (*proto.GenerateUniqueStringResponse, error) {
	uniqueStringGen, ok := g.uniqueStringGenerators[request.SequenceName]
	if !ok {
		strGen, err := gen.NewUniqueString(
			g.dataCollector,
			request.SequenceName,
			int(request.StringLength),
			request.Alphabet,
			g.uniqueNumberGeneratorFactory)
		if err != nil {
			g.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
			return nil, err
		}
		uniqueStringGen = &strGen
		g.uniqueStringGenerators[request.SequenceName] = uniqueStringGen
	}

	uniqueStr, err := uniqueStringGen.GenerateUniqueString()
	if err != nil {
		g.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	return &proto.GenerateUniqueStringResponse{UniqueString: uniqueStr}, nil
}

func NewGenerator(
	dataCollector obs.DataCollector,
	uniqueNumberGeneratorFactory gen.UniqueNumberFactory,
) Generator {
	return Generator{
		dataCollector:                dataCollector,
		uniqueNumberGeneratorFactory: uniqueNumberGeneratorFactory,
		uniqueNumberGenerators:       make(map[string]*gen.UniqueNumber),
		uniqueStringGenerators:       make(map[string]*gen.UniqueString),
	}
}
