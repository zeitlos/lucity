package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	gh "github.com/google/go-github/v92/github"
	"gopkg.in/yaml.v3"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"

	"github.com/zeitlos/lucity/pkg/billing"
	"github.com/zeitlos/lucity/pkg/logto"
	"github.com/zeitlos/lucity/pkg/victoria"
	"github.com/zeitlos/lucity/services/cockpit/internal/cockpit"
)

const (
	platformNamespace = "lucity-system"
	identityService   = "lucity-infra-logto"
	identityPort      = 3001
	metricsService    = "lucity-infra-victoria-metrics-single-server"
	metricsPort       = 8428
)

type secretsFile struct {
	Secrets map[string]string `yaml:"secrets"`
}

type valuesFile struct {
	Conductor struct {
		Env map[string]string `yaml:"env"`
	} `yaml:"conductor"`
}

func connect(ctx context.Context, environment, deployments string) (*cockpit.Client, error) {
	var secrets secretsFile
	var values valuesFile

	directory := filepath.Join(deployments, "lucity-"+environment)

	if err := readYAML(filepath.Join(directory, "secrets.yaml"), &secrets); err != nil {
		return nil, err
	}

	if err := readYAML(filepath.Join(directory, "values.yaml"), &values); err != nil {
		return nil, err
	}

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{CurrentContext: "lucity-" + environment},
	).ClientConfig()

	if err != nil {
		return nil, fmt.Errorf("load kube context lucity-%s: %w", environment, err)
	}

	config.QPS = 50
	config.Burst = 100

	kube, err := kubernetes.NewForConfig(config)

	if err != nil {
		return nil, err
	}

	identityURL, err := forward(ctx, config, kube, identityService, identityPort)

	if err != nil {
		return nil, err
	}

	metricsURL, err := forward(ctx, config, kube, metricsService, metricsPort)

	if err != nil {
		return nil, err
	}

	metrics, err := victoria.New(metricsURL)

	if err != nil {
		return nil, err
	}

	token, err := exec.CommandContext(ctx, "gh", "auth", "token").Output()

	if err != nil {
		return nil, fmt.Errorf("read the GitHub token with gh auth token: %w", err)
	}

	github, err := gh.NewClient(gh.WithAuthToken(strings.TrimSpace(string(token))))

	if err != nil {
		return nil, err
	}

	identity := logto.New(identityURL, secrets.Secrets["LOGTO_M2M_APP_ID"], secrets.Secrets["LOGTO_M2M_APP_SECRET"])
	billingClient := billing.New(secrets.Secrets["STAGEHAND_STRIPE_SECRET_KEY"], values.Conductor.Env["STRIPE_HOBBY_PRICE_ID"], values.Conductor.Env["STRIPE_PRO_PRICE_ID"])

	cacheDirectory, err := os.UserCacheDir()

	if err != nil {
		return nil, err
	}

	return cockpit.New(environment, filepath.Join(cacheDirectory, "lucity-cockpit"), identity, billingClient, kube, metrics, github), nil
}

func readYAML(path string, out any) error {
	raw, err := os.ReadFile(path)

	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}

	return nil
}

func forward(ctx context.Context, config *rest.Config, kube kubernetes.Interface, service string, port int32) (string, error) {
	svc, err := kube.CoreV1().Services(platformNamespace).Get(ctx, service, meta.GetOptions{})

	if err != nil {
		return "", fmt.Errorf("find service %s: %w", service, err)
	}

	pods, err := kube.CoreV1().Pods(platformNamespace).List(ctx, meta.ListOptions{
		LabelSelector: k8slabels.SelectorFromSet(svc.Spec.Selector).String(),
		FieldSelector: "status.phase=Running",
	})

	if err != nil {
		return "", fmt.Errorf("find pods of %s: %w", service, err)
	}

	if len(pods.Items) == 0 {
		return "", fmt.Errorf("no running pod behind %s", service)
	}

	pod := pods.Items[0]
	target := port

	for _, servicePort := range svc.Spec.Ports {
		if servicePort.Port != port {
			continue
		}

		if servicePort.TargetPort.IntValue() > 0 {
			target = int32(servicePort.TargetPort.IntValue())
		}

		for _, container := range pod.Spec.Containers {
			for _, containerPort := range container.Ports {
				if containerPort.Name != "" && containerPort.Name == servicePort.TargetPort.StrVal {
					target = containerPort.ContainerPort
				}
			}
		}
	}

	transport, upgrader, err := spdy.RoundTripperFor(config)

	if err != nil {
		return "", err
	}

	url := kube.CoreV1().RESTClient().Post().Resource("pods").Namespace(platformNamespace).Name(pod.Name).SubResource("portforward").URL()
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, url)

	ready := make(chan struct{})
	forwarder, err := portforward.NewOnAddresses(dialer, []string{"127.0.0.1"}, []string{fmt.Sprintf("0:%d", target)}, ctx.Done(), ready, io.Discard, os.Stderr)

	if err != nil {
		return "", err
	}

	failed := make(chan error, 1)

	go func() { failed <- forwarder.ForwardPorts() }()

	select {
	case <-ready:
	case err := <-failed:
		return "", fmt.Errorf("forward to %s: %w", service, err)
	}

	ports, err := forwarder.GetPorts()

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://127.0.0.1:%d", ports[0].Local), nil
}
