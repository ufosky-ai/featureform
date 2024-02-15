package scheduling

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type TableKey string

func (tk TableKey) ToString() string {
	return string(tk)
}

// Dont like this format that much
func (tk TableKey) GetTaskMetadataKey(id TaskID) string {
	return fmt.Sprintf("%s%d", tk, id)
}

func (tk TableKey) GetTaskRunsKey(id TaskID) string {
	return fmt.Sprintf("%s%d", tk, id)
}

func (tk TableKey) GetTaskRunsMetadataKey() string {
	return fmt.Sprintf("%s", tk)
}

const (
	TASKMETADATA    TableKey = "/tasks/metadata/task_id="
	TASKRUNS        TableKey = "/tasks/runs/task_id="
	TASKRUNMETADATA TableKey = "/tasks/runs/metadata"
)

func NewTaskManager(storage StorageProvider) TaskManager {
	return TaskManager{storage: storage}
}

type TaskManager struct {
	storage StorageProvider
}

type TaskMetadataList []TaskMetadata

func (tml *TaskMetadataList) ToJSON() string {
	return ""
}

// Task Methods
func (tm *TaskManager) CreateTask(name string, tType TaskType, target TaskTarget) (TaskMetadata, error) {
	// ids will be generated by TM
	keys, err := tm.storage.ListKeys(TASKMETADATA.ToString())
	if err != nil {
		return TaskMetadata{}, fmt.Errorf("failed to fetch keys: %v", err)
	}

	var latestID int
	if len(keys) == 0 {
		latestID = 0
	} else {
		latestID, err = getLatestID(keys)
		if err != nil {
			return TaskMetadata{}, err
		}
	}

	metadata := TaskMetadata{
		ID:          TaskID(latestID + 1),
		Name:        name,
		TaskType:    tType,
		Target:      target,
		TargetType:  target.Type(),
		DateCreated: time.Now().UTC(),
	}

	serializedMetadata, err := metadata.Marshal()
	if err != nil {
		return TaskMetadata{}, fmt.Errorf("failed to marshal metadata: %v", err)
	}
	err = tm.storage.Set(TASKMETADATA.GetTaskMetadataKey(metadata.ID), string(serializedMetadata))
	if err != nil {
		return TaskMetadata{}, err
	}

	runs := TaskRuns{
		TaskID: metadata.ID,
		Runs:   []TaskRunSimple{},
	}
	serializedRuns, err := runs.Marshal()
	if err != nil {
		return TaskMetadata{}, err
	}

	err = tm.storage.Set(fmt.Sprintf("/tasks/runs/task_id=%d", metadata.ID), string(serializedRuns))
	if err != nil {
		return TaskMetadata{}, err
	}

	return metadata, nil
}

// Finds the highest increment in a list of strings formatted like "/tasks/metadata/task_id=0"
func getLatestID(taskPaths []string) (int, error) {
	highestIncrement := -1
	for _, path := range taskPaths {
		parts := strings.Split(path, "task_id=")
		if len(parts) < 2 {
			return -1, fmt.Errorf("invalid format for path: %s", path)
		}
		increment, err := strconv.Atoi(parts[1])
		if err != nil {
			return -1, fmt.Errorf("failed to convert task_id to integer: %s", err)
		}
		if increment > highestIncrement {
			highestIncrement = increment
		}
	}
	if highestIncrement == -1 {
		return -1, fmt.Errorf("no valid increments found")
	}
	return highestIncrement, nil
}

func (tm *TaskManager) GetTaskByID(id TaskID) (TaskMetadata, error) {
	key := TASKMETADATA.GetTaskMetadataKey(id)
	metadata, err := tm.storage.Get(key, false)
	if err != nil {
		return TaskMetadata{}, err
	}

	if len(metadata) == 0 {
		return TaskMetadata{}, fmt.Errorf("task not found for id: %s", string(id))
	}

	taskMetadata := TaskMetadata{}
	err = taskMetadata.Unmarshal([]byte(metadata[0]))
	if err != nil {
		return TaskMetadata{}, err
	}
	return taskMetadata, nil
}

func (tm *TaskManager) GetTaskByTarget(target TaskTarget) (TaskMetadataList, error) {
	// need clarification on how to get the task by target
	// and what if the target has multiple tasks?
	// should we return a list of tasks?
	// or do we need to capture uniqueness in the target?
	return []TaskMetadata{}, fmt.Errorf("Not implemented")
}

func (tm *TaskManager) GetAllTasks() (TaskMetadataList, error) {
	// get all the tasks
	metadata, err := tm.storage.Get(TASKMETADATA.ToString(), true)
	if err != nil {
		return TaskMetadataList{}, err
	}

	tml := TaskMetadataList{}
	for _, m := range metadata {
		taskMetadata := TaskMetadata{}
		err = taskMetadata.Unmarshal([]byte(m))
		if err != nil {
			return TaskMetadataList{}, err
		}
		tml = append(tml, taskMetadata)
	}
	return tml, nil
}

type TaskRunList []TaskRunMetadata

func (trl *TaskRunList) ToJSON() string {
	return ""
}

func (trl *TaskRunList) FilterByStatus(status Status) {
	var newList TaskRunList
	for _, run := range *trl {
		if run.Status == status {
			newList = append(newList, run)
		}
	}
	*trl = newList
}

// Task Run Methods
func (tm *TaskManager) CreateTaskRun(name string, taskID TaskID, trigger Trigger) (TaskRunMetadata, error) {
	// ids will be generated by TM
	key, err := tm.storage.Get(fmt.Sprintf("/tasks/runs/task_id=%d", taskID), false)
	if err != nil {
		return TaskRunMetadata{}, fmt.Errorf("failed to fetch task: %v", err)
	}

	runs := TaskRuns{}
	err = runs.Unmarshal([]byte(key[0]))
	if err != nil {
		return TaskRunMetadata{}, err
	}

	latestID, err := getHighestRunID(runs)
	if err != nil {
		return TaskRunMetadata{}, err
	}

	startTime := time.Now().UTC()

	metadata := TaskRunMetadata{
		ID:          TaskRunID(latestID + 1),
		TaskId:      taskID,
		Name:        name,
		Trigger:     trigger,
		TriggerType: trigger.Type(),
		Status:      Pending,
		StartTime:   startTime,
	}

	runs.Runs = append(runs.Runs, TaskRunSimple{RunID: metadata.ID, DateCreated: startTime})

	serializedRuns, err := runs.Marshal()
	if err != nil {
		return TaskRunMetadata{}, err
	}

	serializedMetadata, err := metadata.Marshal()
	if err != nil {
		return TaskRunMetadata{}, fmt.Errorf("failed to marshal metadata: %v", err)
	}
	err = tm.storage.Set(fmt.Sprintf("/tasks/runs/task_id=%d", taskID), string(serializedRuns))
	if err != nil {
		return TaskRunMetadata{}, err
	}

	// Need to double check that date is always 0 padded
	err = tm.storage.Set(fmt.Sprintf("tasks/runs/metadata/%d/%s/%d/task_id=%d/run_id=%d", startTime.Year(), startTime.Month(), startTime.Day(), taskID, metadata.ID), string(serializedMetadata))
	if err != nil {
		return TaskRunMetadata{}, err
	}

	return metadata, nil
}

func getHighestRunID(taskRuns TaskRuns) (TaskRunID, error) {
	if len(taskRuns.Runs) == 0 {
		return 0, nil
	}

	highestRunID := taskRuns.Runs[0].RunID

	for _, run := range taskRuns.Runs[1:] {
		if run.RunID > highestRunID {
			highestRunID = run.RunID
		}
	}

	return highestRunID, nil
}

func (tm *TaskManager) GetRunByID(taskID TaskID, runID TaskRunID) (TaskRunMetadata, error) {
	key, err := tm.storage.Get(fmt.Sprintf("/tasks/runs/task_id=%d", taskID), false)
	if err != nil {
		return TaskRunMetadata{}, fmt.Errorf("failed to fetch task: %v", err)
	}

	runs := TaskRuns{}
	err = runs.Unmarshal([]byte(key[0]))
	if err != nil {
		return TaskRunMetadata{}, err
	}

	found := false
	var runRecord TaskRunSimple
	for _, run := range runs.Runs {
		if run.RunID == runID {
			runRecord = run
			found = true
			break
		}
	}
	if !found {
		return TaskRunMetadata{}, fmt.Errorf("run not found")
	}

	date := runRecord.DateCreated

	rec, err := tm.storage.Get(fmt.Sprintf("tasks/runs/metadata/%d/%s/%d/task_id=%d/run_id=%d", date.Year(), date.Month(), date.Day(), taskID, runRecord.RunID), false)
	if err != nil {
		return TaskRunMetadata{}, err
	}

	taskRun := TaskRunMetadata{}
	err = taskRun.Unmarshal([]byte(rec[0]))
	if err != nil {
		return TaskRunMetadata{}, fmt.Errorf("failed to unmarshal run record: %v", err)
	}
	return taskRun, nil

}

func (tm *TaskManager) GetRunsByDate(start time.Time, end time.Time) (TaskRunList, error) {
	recs, err := tm.storage.Get(fmt.Sprintf("tasks/runs/metadata/%d/%s/%d", start.Year(), start.Month(), start.Day()), true)
	if err != nil {
		return []TaskRunMetadata{}, err
	}

	var runs []TaskRunMetadata
	for _, record := range recs {
		taskRun := TaskRunMetadata{}
		err = taskRun.Unmarshal([]byte(record))
		if err != nil {
			return []TaskRunMetadata{}, fmt.Errorf("failed to unmarshal run record: %v", err)
		}
		if taskRun.StartTime.After(start) {
			continue
		}
		runs = append(runs, taskRun)
	}
	return runs, nil
}

func (tm *TaskManager) GetAllTaskRuns() (TaskRunList, error) {
	recs, err := tm.storage.Get("tasks/runs/metadata", true)
	if err != nil {
		return []TaskRunMetadata{}, err
	}

	var runs []TaskRunMetadata
	for _, record := range recs {
		taskRun := TaskRunMetadata{}
		err = taskRun.Unmarshal([]byte(record))
		if err != nil {
			return []TaskRunMetadata{}, fmt.Errorf("failed to unmarshal run record: %v", err)
		}
		runs = append(runs, taskRun)
	}
	return runs, nil
}

// Write Methods
func (t *TaskManager) SetRunStatus(id TaskRunID, status Status, err error) error {
	// we will need task id as well
	return fmt.Errorf("Not implemented")
}

func (t *TaskManager) SetRunStartTime(id TaskRunID, time time.Time) error {
	// we will need task id as well
	return fmt.Errorf("Not implemented")
}

func (t *TaskManager) SetRunEndTime(id TaskRunID, time time.Time) error {
	// we will need task id as well
	return fmt.Errorf("Not implemented")
}

func (t *TaskManager) AppendRunLog(id TaskRunID, log string) error {
	// we will need task id as well
	return fmt.Errorf("Not implemented")
}

// Locking
func (t *TaskManager) LockTaskRun(ctx context.Context, runId TaskRunID) error {
	// we will need task id as well
	return fmt.Errorf("Not implemented")
}

func (t *TaskManager) UnlockTaskRun(ctx context.Context, runId TaskRunID) error {
	// we will need task id as well
	return fmt.Errorf("Not implemented")
}
