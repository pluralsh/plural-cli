package console

import (
	"encoding/json"
	"strings"

	gqlclient "github.com/pluralsh/console-client-go"
	"github.com/pluralsh/plural-cli/pkg/api"
	"github.com/pluralsh/plural-cli/pkg/utils"
	"github.com/pluralsh/polly/algorithms"
	"github.com/samber/lo"
	batchv1 "k8s.io/api/batch/v1"
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
	Name      string             `json:"name"`
	Type      string             `json:"type"`
	Cluster   string             `json:"cluster"`
	ClusterID string             `json:"clusterId,omitempty"`
	Spec      GateSpecAttributes `json:"spec,omitempty"`
}

type GateSpecAttributes struct {
	Job *GateJobAttributes `json:"job,omitempty"`
}

type GateJobAttributes struct {
	Namespace string `json:"namespace"`
	// if you'd rather define the job spec via straight k8s yaml
	Raw            *batchv1.JobSpec                 `json:"raw,omitempty"`
	Containers     []*gqlclient.ContainerAttributes `json:"containers,omitempty"`
	Labels         *string                          `json:"labels,omitempty"`
	Annotations    *string                          `json:"annotations,omitempty"`
	ServiceAccount *string                          `json:"serviceAccount,omitempty"`
}

func (c *consoleClient) SavePipeline(name string, attrs gqlclient.PipelineAttributes) (*gqlclient.PipelineFragment, error) {
	result, err := c.client.SavePipeline(c.ctx, name, attrs)
	if err != nil {
		return nil, api.GetErrorResponse(err, "SavePipeline")
	}

	return result.SavePipeline, nil
}

func (c *consoleClient) DeletePipeline(id string) (*gqlclient.PipelineFragment, error) {
	result, err := c.client.DeletePipeline(c.ctx, id)
	if err != nil {
		return nil, api.GetErrorResponse(err, "DeletePipeline")
	}

	return result.DeletePipeline, nil
}

func (c *consoleClient) GetPipeline(id string) (*gqlclient.PipelineFragment, error) {
	result, err := c.client.GetPipeline(c.ctx, id)
	if err != nil {
		return nil, api.GetErrorResponse(err, "GetPipeline")
	}

	return result.Pipeline, err
}

func (c *consoleClient) ListPipelines() (*gqlclient.GetPipelines, error) {
	result, err := c.client.GetPipelines(c.ctx, nil)
	if err != nil {
		return nil, api.GetErrorResponse(err, "GetPipelines")
	}
	return result, err
}

func (c *consoleClient) ListPipelineGates() (*gqlclient.GetClusterGates, error) {
	//type GetClusterGates struct {
	//	ClusterGates []*PipelineGateFragment "json:\"clusterGates\" graphql:\"clusterGates\""
	//}
	result, err := c.client.GetClusterGates(c.ctx)
	if err != nil {
		return nil, api.GetErrorResponse(err, "GetClusterGates")
	}
	return result, err
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

//func constructGates(edge PipelineEdge) []*gqlclient.PipelineGateAttributes {
//	res := make([]*gqlclient.PipelineGateAttributes, 0)
//	for _, g := range edge.Gates {
//		res = append(res, &gqlclient.PipelineGateAttributes{
//			Name: g.Name,
//			Type: gqlclient.GateType(strings.ToUpper(g.Type)),
//			//TODO: implement spec parsing
//			Spec: nil,
//		})
//	}
//	return res
//}

func constructGates(edge PipelineEdge) []*gqlclient.PipelineGateAttributes {
	res := make([]*gqlclient.PipelineGateAttributes, 0)
	for _, g := range edge.Gates {
		rawJobSpec, err := json.Marshal(g.Spec.Job.Raw)
		if err != nil {
			utils.Error("unable to marshal raw job spec\n %s", err)
		}
		spec := &gqlclient.GateSpecAttributes{
			Job: &gqlclient.GateJobAttributes{
				Raw:            lo.ToPtr(string(rawJobSpec)),
				Namespace:      g.Spec.Job.Namespace,
				Containers:     g.Spec.Job.Containers,
				Labels:         g.Spec.Job.Labels,
				Annotations:    g.Spec.Job.Annotations,
				ServiceAccount: g.Spec.Job.ServiceAccount,
			},
		}

		res = append(res, &gqlclient.PipelineGateAttributes{
			Name:      g.Name,
			Type:      gqlclient.GateType(strings.ToUpper(g.Type)),
			Cluster:   lo.ToPtr(g.Cluster),
			ClusterID: lo.ToPtr(g.ClusterID),
			Spec:      spec,
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
