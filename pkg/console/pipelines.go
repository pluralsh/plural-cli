package console

import (
	"strings"

	gqlclient "github.com/pluralsh/console/go/client"
	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/polly/algorithms"
	"github.com/samber/lo"
	"sigs.k8s.io/yaml"
)

type Pipeline struct {
	Name   string          `json:"name"`
	Stages []PipelineStage `json:"stages"`
	Edges  []PipelineEdge  `json:"edges"`
}

type PipelineStage struct {
	Name     string         `json:"name"`
	Services []StageService `json:"services"`
}

type StageService struct {
	Name     string             `json:"name"`
	Criteria *PromotionCriteria `json:"criteria"`
}

type PromotionCriteria struct {
	Source  string   `json:"source"`
	Secrets []string `json:"secrets"`
}

type PipelineEdge struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Gates []*Gate
}

type Gate struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Cluster string `json:"cluster"`
}

func (c *consoleClient) SavePipeline(name string, attrs gqlclient.PipelineAttributes) (*gqlclient.PipelineFragment, error) {
	result, err := c.client.SavePipeline(c.ctx, name, attrs)
	if err != nil {
		return nil, api.GetErrorResponse(err, "SavePipeline")
	}

	return result.SavePipeline, nil
}

func ConstructPipelineInput(input []byte) (string, *gqlclient.PipelineAttributes, error) {
	var pipe Pipeline
	if err := yaml.Unmarshal(input, &pipe); err != nil {
		return "", nil, err
	}
	pipeline := &gqlclient.PipelineAttributes{}
	pipeline.Edges = algorithms.Map(pipe.Edges, func(e PipelineEdge) *gqlclient.PipelineEdgeAttributes {
		edge := gqlclient.PipelineEdgeAttributes{From: lo.ToPtr(e.From), To: lo.ToPtr(e.To)}
		edge.Gates = constructGates(e)
		return &edge
	})
	pipeline.Stages = algorithms.Map(pipe.Stages, func(s PipelineStage) *gqlclient.PipelineStageAttributes {
		stage := &gqlclient.PipelineStageAttributes{Name: s.Name}
		stage.Services = algorithms.Map(s.Services, func(s StageService) *gqlclient.StageServiceAttributes {
			handle, name := handleName(s.Name)
			return &gqlclient.StageServiceAttributes{
				Handle:   lo.ToPtr(handle),
				Name:     lo.ToPtr(name),
				Criteria: buildCriteria(s.Criteria),
			}
		})
		return stage
	})
	return pipe.Name, pipeline, nil
}

func constructGates(edge PipelineEdge) []*gqlclient.PipelineGateAttributes {
	res := make([]*gqlclient.PipelineGateAttributes, 0)
	for _, g := range edge.Gates {
		res = append(res, &gqlclient.PipelineGateAttributes{
			Name: g.Name,
			Type: gqlclient.GateType(strings.ToUpper(g.Type)),
		})
	}
	return res
}

func buildCriteria(criteria *PromotionCriteria) *gqlclient.PromotionCriteriaAttributes {
	if criteria == nil {
		return nil
	}

	handle, name := handleName(criteria.Source)
	return &gqlclient.PromotionCriteriaAttributes{
		Handle:  lo.ToPtr(handle),
		Name:    lo.ToPtr(name),
		Secrets: lo.ToSlicePtr(criteria.Secrets),
	}
}

func handleName(name string) (string, string) {
	parts := strings.Split(name, "/")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}

	utils.Error("invalid name: %s, should be of the format {handle}/{name}\n", name)
	return "", ""
}
