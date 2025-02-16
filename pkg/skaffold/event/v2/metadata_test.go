/*
Copyright 2021 The Skaffold Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v2

import (
	"sort"
	"testing"

	latestV1 "github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/latest/v1"
	proto "github.com/GoogleContainerTools/skaffold/proto/v2"
	"github.com/GoogleContainerTools/skaffold/testutil"
)

func TestEmptyState(t *testing.T) {
	tests := []struct {
		description string
		cfg         latestV1.Pipeline
		cluster     string
		expected    *proto.Metadata
	}{
		{
			description: "one build artifact minikube cluster multiple deployers",
			cfg: latestV1.Pipeline{
				Build: latestV1.BuildConfig{
					BuildType: latestV1.BuildType{LocalBuild: &latestV1.LocalBuild{}},
					Artifacts: []*latestV1.Artifact{{ImageName: "docker-artifact-1", ArtifactType: latestV1.ArtifactType{DockerArtifact: &latestV1.DockerArtifact{}}}},
				},
				Deploy: latestV1.DeployConfig{
					DeployType: latestV1.DeployType{
						KubectlDeploy: &latestV1.KubectlDeploy{},
						HelmDeploy:    &latestV1.HelmDeploy{Releases: []latestV1.HelmRelease{{Name: "first"}, {Name: "second"}}},
					},
				},
			},
			cluster: "minikube",
			expected: &proto.Metadata{
				Build: &proto.BuildMetadata{
					Type:      proto.BuildType_LOCAL,
					Artifacts: []*proto.BuildMetadata_Artifact{{Type: proto.BuilderType_DOCKER, Name: "docker-artifact-1"}},
				},
				Deploy: &proto.DeployMetadata{
					Cluster: proto.ClusterType_MINIKUBE,
					Deployers: []*proto.DeployMetadata_Deployer{
						{Type: proto.DeployerType_HELM, Count: 2},
						{Type: proto.DeployerType_KUBECTL, Count: 1},
					},
				},
				RunID: "run-id",
			},
		},
		{
			description: "multiple artifacts of different types gke cluster 1 deployer ",
			cfg: latestV1.Pipeline{
				Build: latestV1.BuildConfig{
					BuildType: latestV1.BuildType{Cluster: &latestV1.ClusterDetails{}},
					Artifacts: []*latestV1.Artifact{
						{ImageName: "docker-artifact-1", ArtifactType: latestV1.ArtifactType{DockerArtifact: &latestV1.DockerArtifact{}}},
						{ImageName: "docker-artifact-2", ArtifactType: latestV1.ArtifactType{DockerArtifact: &latestV1.DockerArtifact{}}},
						{ImageName: "jib-artifact-1", ArtifactType: latestV1.ArtifactType{JibArtifact: &latestV1.JibArtifact{}}},
					},
				},
				Deploy: latestV1.DeployConfig{
					DeployType: latestV1.DeployType{
						KustomizeDeploy: &latestV1.KustomizeDeploy{},
					},
				},
			},
			cluster: "gke-tejal-test",
			expected: &proto.Metadata{
				Build: &proto.BuildMetadata{
					Type: proto.BuildType_CLUSTER,
					Artifacts: []*proto.BuildMetadata_Artifact{
						{Type: proto.BuilderType_JIB, Name: "jib-artifact-1"},
						{Type: proto.BuilderType_DOCKER, Name: "docker-artifact-1"},
						{Type: proto.BuilderType_DOCKER, Name: "docker-artifact-2"},
					},
				},
				Deploy: &proto.DeployMetadata{
					Cluster:   proto.ClusterType_GKE,
					Deployers: []*proto.DeployMetadata_Deployer{{Type: proto.DeployerType_KUSTOMIZE, Count: 1}},
				},
				RunID: "run-id",
			},
		},
		{
			description: "no deployer, kaniko artifact, GCB build",
			cfg: latestV1.Pipeline{
				Build: latestV1.BuildConfig{
					BuildType: latestV1.BuildType{GoogleCloudBuild: &latestV1.GoogleCloudBuild{}},
					Artifacts: []*latestV1.Artifact{
						{ImageName: "artifact-1", ArtifactType: latestV1.ArtifactType{KanikoArtifact: &latestV1.KanikoArtifact{}}},
					},
				},
			},
			cluster: "gke-tejal-test",
			expected: &proto.Metadata{
				Build: &proto.BuildMetadata{
					Type:      proto.BuildType_GCB,
					Artifacts: []*proto.BuildMetadata_Artifact{{Type: proto.BuilderType_KANIKO, Name: "artifact-1"}},
				},
				Deploy: &proto.DeployMetadata{},
				RunID:  "run-id",
			},
		},
		{
			description: "no build, kustomize deployer other cluster",
			cfg: latestV1.Pipeline{
				Deploy: latestV1.DeployConfig{
					DeployType: latestV1.DeployType{
						KustomizeDeploy: &latestV1.KustomizeDeploy{},
					},
				},
			},
			cluster: "some-private",
			expected: &proto.Metadata{
				Build: &proto.BuildMetadata{},
				Deploy: &proto.DeployMetadata{
					Cluster:   proto.ClusterType_OTHER,
					Deployers: []*proto.DeployMetadata_Deployer{{Type: proto.DeployerType_KUSTOMIZE, Count: 1}},
				},
				RunID: "run-id",
			},
		},
	}
	for _, test := range tests {
		testutil.Run(t, test.description, func(t *testutil.T) {
			handler = &eventHandler{
				state: emptyState(mockCfg([]latestV1.Pipeline{test.cfg}, test.cluster)),
			}
			metadata := handler.state.Metadata
			artifacts := metadata.Build.Artifacts

			// sort and compare
			sort.Slice(artifacts, func(i, j int) bool { return artifacts[i].Type < artifacts[j].Type })
			t.CheckDeepEqual(metadata, test.expected)
		})
	}
}
