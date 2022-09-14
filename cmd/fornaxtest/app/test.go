/*
Copyright 2022.

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

package app

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"centaurusinfra.io/fornax-serverless/cmd/fornaxtest/config"
	fornaxv1 "centaurusinfra.io/fornax-serverless/pkg/apis/core/v1"
	fornaxclient "centaurusinfra.io/fornax-serverless/pkg/client/clientset/versioned"
	"centaurusinfra.io/fornax-serverless/pkg/util"
	"github.com/google/uuid"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	_ "k8s.io/component-base/logs/json/register"
	"k8s.io/klog/v2"
)

var (
	allTestSessions = TestSessionArray{}

	apiServerClient = getApiServerClient()

	SessionWrapperEchoServerSpec = &fornaxv1.ApplicationSpec{
		Containers: []v1.Container{{
			Name:  "echoserver",
			Image: "centaurusinfra.io/fornax-serverless/session-wrapper:v0.1.0",
			Ports: []v1.ContainerPort{{
				Name:          "echoserver",
				ContainerPort: 80,
			}},
			Env: []v1.EnvVar{{
				Name:  "SESSION_WRAPPER_OPEN_SESSION_CMD",
				Value: "/opt/bin/sessionwrapper-echoserver",
			}},
			Resources: v1.ResourceRequirements{
				Limits: map[v1.ResourceName]resource.Quantity{
					"memory": util.ResourceQuantity(50*1024*1024, v1.ResourceMemory),
					"cpu":    util.ResourceQuantity(0.05*1000, v1.ResourceCPU),
				},
				Requests: map[v1.ResourceName]resource.Quantity{
					"memory": util.ResourceQuantity(50*1024*1024, v1.ResourceMemory),
					"cpu":    util.ResourceQuantity(0.05*1000, v1.ResourceCPU),
				},
			},
		}},
		ConfigData: map[string]string{},
		ScalingPolicy: fornaxv1.ScalingPolicy{
			MinimumInstance:         0,
			MaximumInstance:         5000,
			Burst:                   50,
			ScalingPolicyType:       "idle_session_number",
			IdleSessionNumThreshold: &fornaxv1.IdelSessionNumThreshold{HighWaterMark: 0, LowWaterMark: 0},
		},
	}

	CloseGracePeriodSeconds             = uint16(10)
	SessionWrapperEchoServerSessionSpec = &fornaxv1.ApplicationSessionSpec{
		ApplicationName:               "",
		SessionData:                   "session-data",
		KillInstanceWhenSessionClosed: false,
		CloseGracePeriodSeconds:       &CloseGracePeriodSeconds,
		OpenTimeoutSeconds:            10,
	}
)

func runAppFullCycleTest(namespace, appName string, testConfig config.TestConfiguration) {
	application, err := describeApplication(namespace, appName)
	if err != nil {
		klog.ErrorS(err, "Failed to find application", "name", appName)
		return
	}
	if application == nil {
		application = createAndWaitForApplicationSetup(namespace, appName, testConfig)
	}

	sessions := createAndWaitForSessionSetup(application, namespace, appName, testConfig)
	cleanupAppFullCycleTest(namespace, appName, sessions)
}

func cleanupAppFullCycleTest(namespace, appName string, sessions []*TestSession) {
	application, _ := describeApplication(namespace, appName)
	instanceNum := application.Status.TotalInstances
	delTime := time.Now()
	deleteApplication(namespace, appName)
	for {
		time.Sleep(200 * time.Millisecond)
		appl, err := describeApplication(namespace, appName)
		if err == nil && appl == nil {
			klog.Infof("Application: %s took %d micro second to teardown %d instances\n", appName, time.Now().Sub(delTime).Microseconds(), instanceNum)
			break
		}
		continue
	}

	for _, v := range sessions {
		deleteSession(v.session.Namespace, v.session.Name)
	}
}

func createAndWaitForApplicationSetup(namespace, appName string, testConfig config.TestConfiguration) *fornaxv1.Application {
	appSpec := SessionWrapperEchoServerSpec.DeepCopy()
	appSpec.ScalingPolicy.Burst = uint32(testConfig.BurstOfPodPerApp)
	appSpec.ScalingPolicy.MinimumInstance = uint32(testConfig.NumOfInitPodsPerApp)
	application, err := createApplication(namespace, appName, appSpec)
	if err != nil {
		klog.ErrorS(err, "Failed to create application", "name", appName)
		return nil
	}
	waitForAppSetup(namespace, appName, int(appSpec.ScalingPolicy.MinimumInstance))
	return application
}

type TestSession struct {
	session            *fornaxv1.ApplicationSession
	creationTimeMicro  int64
	availableTimeMicro int64
	status             fornaxv1.SessionStatus
}

type TestSessionArray []*TestSession

func (sn TestSessionArray) Len() int {
	return len(sn)
}

//so, sort latency from smaller to lager value
func (sn TestSessionArray) Less(i, j int) bool {
	return sn[i].availableTimeMicro-sn[i].creationTimeMicro < sn[j].availableTimeMicro-sn[j].creationTimeMicro
}

func (sn TestSessionArray) Swap(i, j int) {
	sn[i], sn[j] = sn[j], sn[i]
}
func createAndWaitForSessionSetup(application *fornaxv1.Application, namespace, appName string, testConfig config.TestConfiguration) []*TestSession {
	sessions := []*TestSession{}
	sessionBaseName := uuid.New().String()
	numOfSession := testConfig.NumOfSessionPerApp
	staggerNum := numOfSession / 5
	for i := 0; i < numOfSession; i++ {
		sessName := fmt.Sprintf("%s-%s-session-%d", appName, sessionBaseName, i)
		ts, _ := createSession(namespace, sessName, appName, SessionWrapperEchoServerSessionSpec)
		sessions = append(sessions, ts)
		// create sessions in 5 batchs, every batch interval 200 milli seconds, we want to control rate in a second
		if i > 0 && i%staggerNum == 0 {
			time.Sleep(200 * time.Millisecond)
		}
	}

	waitForSessionSetup(namespace, appName, sessions)
	return sessions
}

func waitForAppSetup(namespace, appName string, numOfInstance int) {
	if numOfInstance > 0 {
		klog.Infof("waiting for %d pods of app %s setup", numOfInstance, appName)
		for {
			time.Sleep(200 * time.Millisecond)
			application, err := describeApplication(namespace, appName)
			if err != nil {
				continue
			}

			if int(application.Status.RunningInstances) >= numOfInstance {
				ct := application.CreationTimestamp.UnixMicro()
				if v, found := application.Labels[fornaxv1.LabelFornaxCoreCreationUnixMicro]; found {
					t, _ := strconv.Atoi(v)
					ct = int64(t)
				}
				at := time.Now().UnixMicro()
				klog.Infof("Application: %s took %d micro second to setup %d instances\n", appName, at-ct, application.Status.RunningInstances)
				break
			}
			continue
		}
	}
}

func waitForSessionTearDown(namespace, appName string, sessions []*TestSession) {
	if len(sessions) > 0 {
		klog.Infof("waiting for %d sessions of app %s teardown", len(sessions), appName)
		for {
			time.Sleep(500 * time.Millisecond)
			app, err := describeApplication(namespace, appName)
			if err != nil {
				continue
			}

			if app == nil || app.Status.RunningInstances-app.Status.IdleInstances == 0 {
				// all instance are release or recreated
				break
			}

			// allTeardown := true
			// for _, ts := range sessions {
			//  if ts.status == fornaxv1.SessionStatusClosed {
			//    continue
			//  }
			//  sess, err := describeSession(getApiServerClient(), namespace, ts.session.Name)
			//
			//  if err != nil || sess != nil {
			//    allTeardown = false
			//  }
			//
			//  if sess == nil {
			//    ts.status = fornaxv1.SessionStatusClosed
			//  }
			// }
			//
			// if allTeardown {
			//  break
			// }
		}
	}
}

func waitForSessionSetup(namespace, appName string, sessions TestSessionArray) {
	if len(sessions) > 0 {
		klog.Infof("waiting for %d sessions of app %s setup", len(sessions), appName)
		for {
			time.Sleep(200 * time.Millisecond)
			allSetup := true
			for _, ts := range sessions {
				if ts.status != fornaxv1.SessionStatusPending {
					continue
				}
				sess, err := describeSession(namespace, ts.session.Name)
				if err != nil || sess == nil {
					allSetup = false
					continue
				}

				switch sess.Status.SessionStatus {
				case fornaxv1.SessionStatusAvailable:
					ts.availableTimeMicro = sess.Status.AvailableTimeMicro
					// ts.availableTimeMicro = time.Now().UnixMicro()
					ts.status = sess.Status.SessionStatus
				case fornaxv1.SessionStatusTimeout:
					ts.status = sess.Status.SessionStatus
				case fornaxv1.SessionStatusClosed:
					ts.status = sess.Status.SessionStatus
				default:
					allSetup = false
					continue
				}
			}

			if allSetup {
				// for _, v := range sessions {
				//  klog.Infof("Session: %s took %d micro second to setup, status %s\n", v.session.Name, v.availableTimeMicro-v.creationTimeMicro, v.status)
				// }
				break
			}
		}

		allTestSessions = append(allTestSessions, sessions...)
	}
}

func createSessionTest(namespace, appName string, testConfig config.TestConfiguration) {
	application, err := describeApplication(namespace, appName)
	if err != nil {
		klog.ErrorS(err, "Failed to find application", "name", appName)
		return
	}
	if application == nil {
		application = createAndWaitForApplicationSetup(namespace, appName, testConfig)
	}

	createAndWaitForSessionSetup(application, namespace, appName, testConfig)
}

func runSessionFullCycleTest(namespace, appName string, testConfig config.TestConfiguration) {
	application, err := describeApplication(namespace, appName)
	if err != nil {
		klog.ErrorS(err, "Failed to find application", "name", appName)
		return
	}
	if application == nil {
		application = createAndWaitForApplicationSetup(namespace, appName, testConfig)
	}

	sessions := createAndWaitForSessionSetup(application, namespace, appName, testConfig)
	cleanupSessionFullCycleTest(namespace, appName, sessions)
}

func cleanupSessionFullCycleTest(namespace, appName string, sessions []*TestSession) {
	for _, sess := range sessions {
		go deleteSession(namespace, sess.session.Name)
	}
	waitForSessionTearDown(namespace, appName, sessions)
}

func createApplication(namespace, name string, appSpec *fornaxv1.ApplicationSpec) (*fornaxv1.Application, error) {
	client := getApiServerClient()
	appClient := client.CoreV1().Applications(namespace)
	application := &fornaxv1.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       fornaxv1.ApplicationKind.Kind,
			APIVersion: fornaxv1.ApplicationKind.Version,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:         name,
			GenerateName: name,
			Namespace:    namespace,
			Labels: map[string]string{
				fornaxv1.LabelFornaxCoreCreationUnixMicro: fmt.Sprint(time.Now().UnixMicro()),
			},
		},
		Spec:   *appSpec.DeepCopy(),
		Status: fornaxv1.ApplicationStatus{},
	}
	klog.InfoS("Application created", "application", util.Name(application), "initial pods", application.Spec.ScalingPolicy.MinimumInstance)
	return appClient.Create(context.Background(), application, metav1.CreateOptions{})
}

func createSession(namespace, name, applicationName string, sessionSpec *fornaxv1.ApplicationSessionSpec) (*TestSession, error) {
	client := getApiServerClient()
	appClient := client.CoreV1().ApplicationSessions(namespace)
	spec := *sessionSpec.DeepCopy()
	spec.ApplicationName = applicationName
	session := &fornaxv1.ApplicationSession{
		TypeMeta: metav1.TypeMeta{
			Kind:       fornaxv1.ApplicationSessionKind.Kind,
			APIVersion: fornaxv1.ApplicationSessionKind.Version,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:         name,
			GenerateName: name,
			Namespace:    namespace,
			Labels: map[string]string{
				fornaxv1.LabelFornaxCoreCreationUnixMicro: fmt.Sprint(time.Now().UnixMicro()),
			},
		},
		Spec: spec,
	}
	creationTimeMicro := time.Now().UnixMicro()
	session, err := appClient.Create(context.Background(), session, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	// klog.InfoS("Session created", "application", applicationName, "session", name)
	return &TestSession{
		session:            session,
		creationTimeMicro:  creationTimeMicro,
		availableTimeMicro: 0,
		status:             fornaxv1.SessionStatusPending,
	}, nil
}

func deleteApplication(namespace, name string) error {
	client := getApiServerClient()
	appClient := client.CoreV1().Applications(namespace)
	klog.InfoS("Applications deleted", "namespace", namespace, "app", name)
	return appClient.Delete(context.Background(), name, metav1.DeleteOptions{})
}

func deleteSession(namespace, name string) error {
	client := getApiServerClient()
	appClient := client.CoreV1().ApplicationSessions(namespace)
	return appClient.Delete(context.Background(), name, metav1.DeleteOptions{})
}

func describeApplication(namespace, name string) (*fornaxv1.Application, error) {
	client := getApiServerClient()
	appClient := client.CoreV1().Applications(namespace)
	apps, err := appClient.Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return apps, err
}
func describeSession(namespace, name string) (*fornaxv1.ApplicationSession, error) {
	client := getApiServerClient()
	appClient := client.CoreV1().ApplicationSessions(namespace)
	sess, err := appClient.Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return sess, err
}

func getApiServerClient() *fornaxclient.Clientset {
	if root, err := os.Getwd(); err == nil {
		kubeconfigPath := root + "/kubeconfig"
		if kubeconfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath); err != nil {
			klog.ErrorS(err, "Failed to construct kube rest config")
			os.Exit(-1)
		} else {
			return fornaxclient.NewForConfigOrDie(kubeconfig)
		}
	} else {
		klog.ErrorS(err, "Failed to get working dir")
		os.Exit(-1)
	}

	return nil
}

func summarySessionTestResult(sessions TestSessionArray, startTimeMilli, endTimeMilli int64) {
	failedSession := 0
	timeoutSession := 0
	successSession := 0
	for _, v := range sessions {
		if v.status == fornaxv1.SessionStatusClosed {
			failedSession += 1
		}
		if v.status == fornaxv1.SessionStatusAvailable {
			successSession += 1
		}
		if v.status == fornaxv1.SessionStatusTimeout {
			timeoutSession += 1
		}
	}
	if len(sessions) == 0 {
		return
	}
	klog.Infof("--------%d Session created----------\n", len(sessions))
	klog.Infof("Test use %d ms", (endTimeMilli - startTimeMilli))
	klog.Infof("Session creation rate %d/s", int64(len(sessions))*1000/(endTimeMilli-startTimeMilli))
	klog.Infof("%d success, %d failed, %d timeout", successSession, failedSession, timeoutSession)
	sort.Sort(sessions)
	p99 := sessions[len(sessions)*99/100]
	p90 := sessions[len(sessions)*90/100]
	p50 := sessions[len(sessions)*50/100]
	klog.Infof("Session setup time: p99 %d micro seconds, %v", p99.availableTimeMicro-p99.creationTimeMicro, p99)
	klog.Infof("Session setup time: p90 %d micro seconds", p90.availableTimeMicro-p90.creationTimeMicro)
	klog.Infof("Session setup time: p50 %d micro seconds", p50.availableTimeMicro-p50.creationTimeMicro)
}
