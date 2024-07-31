package main

import (
	"context"
	"fmt"
	mod "mem_monitor/modules"
	"os"
	"strconv"
	"time"

	"github.com/shirou/gopsutil/mem"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	internalapi "k8s.io/cri-api/pkg/apis"
	pb "k8s.io/cri-api/pkg/apis/runtime/v1"
)

func GetSystemMemoryStatsInfo() mod.Memory {
	var get_memory mod.Memory

	memory, err := mem.VirtualMemory()
	if err != nil {
		panic(err)
	}
	// fmt.Println(memory)

	get_memory.Total = memory.Total
	get_memory.Available = memory.Available
	get_memory.Used = memory.Total - memory.Available
	get_memory.UsedPercent = float64(get_memory.Used) / float64(memory.Total) * 100

	return get_memory
}

func GetPodInfo(client internalapi.RuntimeService) bool {

	var count int64 = 0

	filter := &pb.PodSandboxStatsFilter{}

	stats, err := client.ListPodSandboxStats(context.TODO(), filter)
	if err != nil {
		fmt.Println(err)
		return true
	}

	for i := 0; i < len(stats); i++ {
		// Do not store namespaces other than default namespaces
		if stats[i].Attributes.Metadata.Namespace != "default" {
			continue
		}
		// Do not store info of notworking pods
		status, err := client.PodSandboxStatus(context.TODO(), stats[i].Attributes.Id, false)
		if err != nil {
			fmt.Println(err)
		}
		if status.Status.State == 1 { // exception handling: SANDBOX_NOTREADY
			continue
		}
		count++
	}

	if count == 0 {
		return false
	} else {
		return true
	}
}

type PodData struct {
	PodName      string
	StartTime    int64
	StartedAt    int64
	FinishedAt   int64
	RunningTime  int64
	WaitTime     int64
	RestartCount int32
}

func IsSucceed(podsItems []v1.Pod) bool {
	for _, pod := range podsItems {
		if pod.Status.Phase != "Succeeded" {
			return false
		}
	}
	return true
}

func main() {
	var pods *v1.PodList
	var err error

	var podData []PodData
	var runningTimeArr []int64
	var waitTimeArr []int64

	var numOfWorkers int64 = 3

	var startedTestTime int64 = 9999999999999
	var finishedTestTime int64 = 0

	var minContainerRunningTime int64 = 9999999999999
	var maxContainerRunningTime int64 = 0
	var totalContinerRunningTime int64 = 0
	var totalContainerRestart int64 = 0
	var containerRestartArray [20]int64

	var minContainerWaitTime int64 = 9999999999999
	var maxContainerWaitTime int64 = 0
	var totalContainerWaitTime int64 = 0

	// kubernetes api 클라이언트 생성하는 모듈
	clientset := mod.InitClient()
	if clientset == nil {
		fmt.Println("Could not create client!")
		os.Exit(-1)
	}

	for {
		pods, err = clientset.CoreV1().Pods("default").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err)
		}

		if IsSucceed(pods.Items) == false {
			time.Sleep(time.Second)
		} else {
			break
		}
	}

	for _, pod := range pods.Items {
		var newPod PodData

		newPod.PodName = pod.Name

		newPod.StartTime = pod.Status.StartTime.Unix()
		newPod.StartedAt = pod.Status.ContainerStatuses[0].State.Terminated.StartedAt.Unix()
		newPod.FinishedAt = pod.Status.ContainerStatuses[0].State.Terminated.FinishedAt.Unix()
		newPod.RunningTime = newPod.FinishedAt - newPod.StartTime
		//newPod.WaitTime = newPod.StartedAt - newPod.StartTime
		newPod.RestartCount = pod.Status.ContainerStatuses[0].RestartCount

		/*fmt.Println()
		fmt.Println(newPod.PodName)

		fmt.Println("PodContitions")
		for i := 0; i < len(pod.Status.Conditions); i++ {
			fmt.Println(pod.Status.Conditions[i].Type, ":", pod.Status.Conditions[i].LastTransitionTime.Unix())
		}
		fmt.Println("Time Info")
		fmt.Println("StartTime:", newPod.StartTime)
		fmt.Println("StartedAt:", newPod.StartedAt)
		fmt.Println("FinishedAt:", newPod.FinishedAt)

		fmt.Println("running Time:", newPod.RunningTime)
		fmt.Println("RestartCount:", newPod.RestartCount)*/

		if startedTestTime > newPod.StartTime {
			startedTestTime = newPod.StartTime
		}
		if finishedTestTime < newPod.FinishedAt {
			finishedTestTime = newPod.FinishedAt
		}

		if minContainerRunningTime > newPod.RunningTime {
			minContainerRunningTime = newPod.RunningTime
		}
		if maxContainerRunningTime < newPod.RunningTime {
			maxContainerRunningTime = newPod.RunningTime
		}

		totalContinerRunningTime += newPod.RunningTime
		totalContainerRestart += int64(pod.Status.ContainerStatuses[0].RestartCount)
		containerRestartArray[pod.Status.ContainerStatuses[0].RestartCount]++

		podData = append(podData, newPod)
		//a := pod.Status.
	}

	for i, pod := range podData {
		var bias int64
		var err error
		bias, err = strconv.ParseInt(pod.PodName[len(pod.PodName)-3:], 10, 64)
		if err != nil {
			bias, err = strconv.ParseInt(pod.PodName[len(pod.PodName)-2:], 10, 64)
			if err != nil {
				bias, _ = strconv.ParseInt(pod.PodName[len(pod.PodName)-1:], 10, 64)
			}
		}
		podData[i].WaitTime = pod.StartedAt - startedTestTime - (bias / numOfWorkers)

		if minContainerWaitTime > pod.WaitTime {
			minContainerWaitTime = pod.WaitTime
		}
		if maxContainerWaitTime < pod.WaitTime {
			maxContainerWaitTime = pod.WaitTime
		}
		totalContainerWaitTime += pod.WaitTime
	}

	for _, pod := range podData {
		runningTimeArr = append(runningTimeArr, pod.RunningTime)
		waitTimeArr = append(waitTimeArr, pod.WaitTime)

	}

	//print total info
	fmt.Println()
	fmt.Println("Total Info")
	fmt.Println("total TestingTime:", finishedTestTime-startedTestTime)
	//fmt.Println("minContainerRunningTime:", minContainerRunningTime)
	fmt.Println("averageContainerRunningTime:", float64(totalContinerRunningTime)/float64(len(pods.Items)))
	//fmt.Println("maxContainerRunningTime:", maxContainerRunningTime)
	fmt.Println("containerRunningTimeArray:", runningTimeArr)
	fmt.Println("totalContainerRestart:", totalContainerRestart)
	fmt.Println("containerRestartArray:", containerRestartArray)
	fmt.Println("containerRestartRatio:", float64(totalContainerRestart)/float64(len(podData)))
	//fmt.Println("minContainerWaitTime:", minContainerWaitTime)
	fmt.Println("averageContainerWaitTime:", float64(totalContainerWaitTime)/float64(len(pods.Items)))
	//fmt.Println("maxContainerWaitTime:", maxContainerWaitTime)
	fmt.Println("containerWaitTimeArray:", waitTimeArr)

}
