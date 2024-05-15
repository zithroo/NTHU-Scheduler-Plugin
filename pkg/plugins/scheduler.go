package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

type CustomSchedulerArgs struct {
	Mode string `json:"mode"`
}

type CustomScheduler struct {
	handle    framework.Handle
	scoreMode string
}

var _ framework.PreFilterPlugin = &CustomScheduler{}
var _ framework.ScorePlugin = &CustomScheduler{}

// Name is the name of the plugin used in Registry and configurations.
const (
	Name              string = "CustomScheduler"
	groupNameLabel    string = "podGroup"
	minAvailableLabel string = "minAvailable"
	leastMode         string = "Least"
	mostMode          string = "Most"
)

func (cs *CustomScheduler) Name() string {
	return Name
}

// New initializes and returns a new CustomScheduler plugin.
func New(obj runtime.Object, h framework.Handle) (framework.Plugin, error) {
	cs := CustomScheduler{}
	mode := leastMode
	if obj != nil {
		args := obj.(*runtime.Unknown)
		var csArgs CustomSchedulerArgs
		if err := json.Unmarshal(args.Raw, &csArgs); err != nil {
			fmt.Printf("Error unmarshal: %v\n", err)
		}
		mode = csArgs.Mode
		if mode != leastMode && mode != mostMode {
			return nil, fmt.Errorf("invalid mode, got %s", mode)
		}
	}
	cs.handle = h
	cs.scoreMode = mode
	log.Printf("Custom scheduler runs with the mode: %s.", mode)

	return &cs, nil
}

// filter the pod if the pod in group is less than minAvailable
func (cs *CustomScheduler) PreFilter(ctx context.Context, state *framework.CycleState, pod *v1.Pod) (*framework.PreFilterResult, *framework.Status) {
	log.Printf("Pod %s is in Prefilter phase.", pod.Name)
	newStatus := framework.NewStatus(framework.Success, "")

	// TODO
	// 1. extract the label of the pod
	// 2. retrieve the pod with the same group label
	// 3. justify if the pod can be scheduled

	podLables := pod.ObjectMeta.GetLabels()
	selector := labels.SelectorFromSet(map[string]string{groupNameLabel: podLables[groupNameLabel]})
	pods, _ := cs.handle.SharedInformerFactory().Core().V1().Pods().Lister().List(selector)
	minAvailable, err := strconv.Atoi(podLables[minAvailableLabel])
	if err != nil || len(pods) < minAvailable {
		return nil, framework.NewStatus(framework.Unschedulable, "Not enough pods in group")
	}
	return nil, newStatus
}

// PreFilterExtensions returns a PreFilterExtensions interface if the plugin implements one.
func (cs *CustomScheduler) PreFilterExtensions() framework.PreFilterExtensions {
	return nil
}

// Score invoked at the score extension point.
func (cs *CustomScheduler) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	log.Printf("Pod %s is in Score phase. Calculate the score of Node %s.", pod.Name, nodeName)

	// TODO
	// 1. retrieve the node allocatable memory
	// 2. return the score based on the scheduler mode

	nodeInfo, _ := cs.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	allocatableMemory := nodeInfo.Allocatable.Memory

	if cs.scoreMode == leastMode {
		return -int64(allocatableMemory), nil
	} else if cs.scoreMode == mostMode {
		return int64(allocatableMemory), nil
	} else {
		return 0, nil
	}
}

// ensure the scores are within the valid range
func (cs *CustomScheduler) NormalizeScore(ctx context.Context, state *framework.CycleState, pod *v1.Pod, scores framework.NodeScoreList) *framework.Status {
	// TODO
	// find the range of the current score and map to the valid range

	// map to 0~100
	var maxScore int64 = math.MinInt64
	var minScore int64 = math.MaxInt64
	for i := range scores {
		if scores[i].Score > maxScore {
			maxScore = scores[i].Score
		}
		if scores[i].Score < minScore {
			minScore = scores[i].Score
		}
	}

	for i := range scores {
		scores[i].Score = (scores[i].Score - minScore) * 100 / (maxScore - minScore)
	}
	return nil
}

// ScoreExtensions of the Score plugin.
func (cs *CustomScheduler) ScoreExtensions() framework.ScoreExtensions {
	return cs
}
