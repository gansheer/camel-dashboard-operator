/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package jib contains utilities for jib strategy builds.
package jib

import (
	"context"
	"encoding/xml"
	"fmt"

	v1 "github.com/squakez/camel-dashboard-operator/pkg/apis/camel/v1"
	"github.com/squakez/camel-dashboard-operator/pkg/client"
	"github.com/squakez/camel-dashboard-operator/pkg/util"
	"github.com/squakez/camel-dashboard-operator/pkg/util/maven"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const JibMavenGoal = "jib:build"
const JibMavenToImageParam = "-Djib.to.image="
const JibMavenFromImageParam = "-Djib.from.image="
const JibMavenFromPlatforms = "-Djib.from.platforms="
const JibMavenBaseImageCache = "-Djib.baseImageCache="
const JibMavenInsecureRegistries = "-Djib.allowInsecureRegistries="
const JibDigestFile = "target/jib-image.digest"
const JibMavenPluginVersionDefault = "3.4.1"
const JibLayerFilterExtensionMavenVersionDefault = "0.3.0"

// See: https://github.com/GoogleContainerTools/jib/blob/master/jib-maven-plugin/README.md#using-docker-configuration-files
const JibRegistryConfigEnvVar = "DOCKER_CONFIG"

// The Jib profile configuration.
const XMLJibProfile = `
<profile>
  <id>jib</id>
  <activation>
    <activeByDefault>false</activeByDefault>
  </activation>
  <repositories></repositories>
  <pluginRepositories></pluginRepositories>
  <build>
    <plugins>
      <plugin>
        <groupId>com.google.cloud.tools</groupId>
        <artifactId>jib-maven-plugin</artifactId>
        <version>3.4.1</version>
        <executions></executions>
        <dependencies>
          <dependency>
            <groupId>com.google.cloud.tools</groupId>
            <artifactId>jib-layer-filter-extension-maven</artifactId>
            <version>0.3.0</version>
          </dependency>
        </dependencies>
        <configuration>
          <container>
            <entrypoint>INHERIT</entrypoint>
            <args>
              <arg>jshell</arg>
            </args>
          </container>
          <allowInsecureRegistries>true</allowInsecureRegistries>
          <extraDirectories>
            <paths>
              <path>
                <from>../context</from>
                <into>/deployments</into>
                <excludes></excludes>
              </path>
            </paths>
            <permissions>
              <permission>
                <file>/deployments/*</file>
                <mode>755</mode>
              </permission>
            </permissions>
          </extraDirectories>
          <pluginExtensions>
            <pluginExtension>
              <implementation>com.google.cloud.tools.jib.maven.extension.layerfilter.JibLayerFilterExtension</implementation>
              <configuration implementation="com.google.cloud.tools.jib.maven.extension.layerfilter.Configuration">
                <filters>
                  <Filter>
                    <glob>/app/**</glob>
                  </Filter>
                </filters>
              </configuration>
            </pluginExtension>
          </pluginExtensions>
        </configuration>
      </plugin>
    </plugins>
  </build>
</profile>
`

type JibBuild struct {
	Plugins []maven.Plugin `xml:"plugins>plugin,omitempty"`
}

type JibProfile struct {
	XMLName xml.Name
	ID      string   `xml:"id"`
	Build   JibBuild `xml:"build,omitempty"`
}

// Create a Configmap containing the default jib profile.
func CreateProfileConfigmap(ctx context.Context, c client.Client,
	profile, name, namespace, apiVersion, kind, label string, annotations map[string]string, UID types.UID) error {
	controller := true
	blockOwnerDeletion := true
	jibProfileConfigMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name + "-publish-jib-profile",
			Namespace:   namespace,
			Annotations: annotations,
			Labels: map[string]string{
				label: name,
			},
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         apiVersion,
					Kind:               kind,
					Name:               name,
					UID:                UID,
					Controller:         &controller,
					BlockOwnerDeletion: &blockOwnerDeletion,
				}},
		},
		Data: map[string]string{
			"profile.xml": profile,
		},
	}

	err := c.Create(ctx, jibProfileConfigMap)
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return fmt.Errorf("error creating the configmap containing the default maven jib profile: %s-publish-jib-profile: %w", name, err)
	}
	return nil
}

// JibMavenProfile creates a maven profile defining jib plugin build.
func JibMavenProfile(jibMavenPluginVersion string, jibLayerFilterExtensionMavenVersion string) (string, error) {
	jibVersion := JibMavenPluginVersionDefault
	if jibMavenPluginVersion != "" {
		jibVersion = jibMavenPluginVersion
	}
	layerVersion := JibLayerFilterExtensionMavenVersionDefault
	if jibLayerFilterExtensionMavenVersion != "" {
		layerVersion = jibLayerFilterExtensionMavenVersion
	}
	jibPlugin := maven.Plugin{
		GroupID:    "com.google.cloud.tools",
		ArtifactID: "jib-maven-plugin",
		Version:    jibVersion,
		Dependencies: []maven.Dependency{
			{
				GroupID:    "com.google.cloud.tools",
				ArtifactID: "jib-layer-filter-extension-maven",
				Version:    layerVersion,
			},
		},
		Configuration: v1.PluginConfiguration{
			Container: v1.Container{
				Entrypoint: "INHERIT",
				Args: v1.Args{
					Arg: "jshell",
				},
			},
			AllowInsecureRegistries: "true",
			ExtraDirectories: v1.ExtraDirectories{
				Paths: []v1.Path{
					{
						From: "../context",
						Into: "/deployments",
					},
				},
				Permissions: []v1.Permission{
					{
						File: "/deployments/*",
						Mode: "755",
					},
				},
			},
			PluginExtensions: v1.PluginExtensions{
				PluginExtension: v1.PluginExtension{
					Implementation: "com.google.cloud.tools.jib.maven.extension.layerfilter.JibLayerFilterExtension",
					Configuration: v1.PluginExtensionConfiguration{
						Implementation: "com.google.cloud.tools.jib.maven.extension.layerfilter.Configuration",
						Filters: []v1.Filter{
							{
								Glob: "/app/**",
							},
						},
					},
				},
			},
		},
	}

	jibMavenPluginProfile := JibProfile{
		XMLName: xml.Name{Local: "profile"},
		ID:      "jib",
		Build: JibBuild{
			Plugins: []maven.Plugin{jibPlugin},
		},
	}
	content, err := util.EncodeXMLWithoutHeader(jibMavenPluginProfile)
	if err != nil {
		return "", err
	}
	return string(content), nil

}
