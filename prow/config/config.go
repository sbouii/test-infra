/*
Copyright 2017 The Kubernetes Authors.

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

// Package config knows how to read and parse config.yaml.
package config

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"text/template"
	"time"

	"github.com/ghodss/yaml"
)

// Config is a read-only snapshot of the config.
type Config struct {
	// Full repo name (such as "kubernetes/kubernetes") -> list of jobs.
	Presubmits  map[string][]Presubmit  `json:"presubmits,omitempty"`
	Postsubmits map[string][]Postsubmit `json:"postsubmits,omitempty"`

	// Periodics are not associated with any repo.
	Periodics []Periodic `json:"periodics,omitempty"`

	Plank    Plank     `json:"plank,omitempty"`
	Triggers []Trigger `json:"triggers,omitempty"`

	// ProwJobNamespace is the namespace in the cluster that prow
	// components will use for looking up ProwJobs. The namespace
	// needs to exist and will not be created by prow.
	ProwJobNamespace string `json:"prowjob_namespace,omitempty"`
	// PodNamespace is the namespace in the cluster that prow
	// components will use for looking up Pods owned by ProwJobs.
	// The namespace needs to exist and will not be created by prow.
	PodNamespace string `json:"pod_namespace,omitempty"`
}

// Plank is config for the plank controller.
type Plank struct {
	// JobURLTemplateString compiles into JobURLTemplate at load time.
	JobURLTemplateString string `json:"job_url_template,omitempty"`
	// JobURLTemplate is compiled at load time from JobURLTemplateString. It
	// will be passed a kube.ProwJob and is used to set the URL for the
	// "details" link on GitHub as well as the link from deck.
	JobURLTemplate *template.Template `json:"-"`

	// ReportTemplateString compiles into ReportTemplate at load time.
	ReportTemplateString string `json:"report_template,omitempty"`
	// ReportTemplate is compiled at load time from ReportTemplateString. It
	// will be passed a kube.ProwJob and can provide an optional blurb below
	// the test failures comment.
	ReportTemplate *template.Template `json:"-"`
}

// Trigger is config for the trigger plugin.
type Trigger struct {
	// Repos is either of the form org/repos or just org.
	Repos []string `json:"repos,omitempty"`
	// TrustedOrg is the org whose members' PRs will be automatically built
	// for PRs to the above repos.
	TrustedOrg string `json:"trusted_org,omitempty"`
}

// Load loads and parses the config at path.
func Load(path string) (*Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %v", path, err)
	}
	nc := &Config{}
	if err := yaml.Unmarshal(b, nc); err != nil {
		return nil, fmt.Errorf("error unmarshaling %s: %v", path, err)
	}
	if err := parseConfig(nc); err != nil {
		return nil, err
	}
	return nc, nil
}

func parseConfig(c *Config) error {
	// Ensure that presubmit regexes are valid.
	for _, v := range c.Presubmits {
		if err := setRegexes(v); err != nil {
			return fmt.Errorf("could not set regex: %v", err)
		}
	}

	// Ensure that postsubmits have a pod spec.
	for _, js := range c.Postsubmits {
		for j := range js {
			if js[j].Spec == nil {
				return fmt.Errorf("job %s has no spec", js[j].Name)
			}
		}
	}

	// Ensure that the periodic durations are valid and specs exist.
	for j := range c.Periodics {
		if c.Periodics[j].Spec == nil {
			return fmt.Errorf("job %s has no spec", c.Periodics[j].Name)
		}
		d, err := time.ParseDuration(c.Periodics[j].Interval)
		if err != nil {
			return fmt.Errorf("cannot parse duration for %s: %v", c.Periodics[j].Name, err)
		}
		c.Periodics[j].interval = d
	}

	urlTmpl, err := template.New("JobURL").Parse(c.Plank.JobURLTemplateString)
	if err != nil {
		return fmt.Errorf("parsing template: %v", err)
	}
	c.Plank.JobURLTemplate = urlTmpl

	reportTmpl, err := template.New("Report").Parse(c.Plank.ReportTemplateString)
	if err != nil {
		return fmt.Errorf("parsing template: %v", err)
	}
	c.Plank.ReportTemplate = reportTmpl
	return nil
}

func setRegexes(js []Presubmit) error {
	for i, j := range js {
		if re, err := regexp.Compile(j.Trigger); err == nil {
			js[i].re = re
		} else {
			return fmt.Errorf("could not compile trigger regex for %s: %v", j.Name, err)
		}
		if err := setRegexes(j.RunAfterSuccess); err != nil {
			return err
		}
		if j.RunIfChanged != "" {
			re, err := regexp.Compile(j.RunIfChanged)
			if err != nil {
				return fmt.Errorf("could not compile changes regex for %s: %v", j.Name, err)
			}
			js[i].reChanges = re
		}
	}
	return nil
}
