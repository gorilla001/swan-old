package mesos

import (
	//"errors"
	//"fmt"
	//"strings"
	//"time"

	"github.com/Dataman-Cloud/swan/types"
	//log "github.com/Sirupsen/logrus"
)

type AppMode string

var (
	APP_MODE_FIXED      AppMode = "fixed"
	APP_MODE_REPLICATES AppMode = "replicates"
)

//func (s *Scheduler) CreateApp(version *types.Version, insName string) (*types.Application, error) {
//	if insName == "" {
//		insName = "default"
//	}
//
//	var (
//		v       = version
//		id      = fmt.Sprintf("%s.%s.%s.%s", v.AppName, insName, v.RunAs, s.ClusterName())
//		network = strings.ToLower(version.Container.Docker.Network)
//		count   = int(v.Instances)
//	)
//
//	if app := s.db.GetApp(id); app != nil {
//		return nil, errors.New("app already exists")
//	}
//
//	app := &types.Application{
//		ID:        id,
//		Name:      v.AppName,
//		Version:   v,
//		RunAs:     v.RunAs,
//		ClusterID: s.ClusterName(),
//		Status:    "creating",
//		CreatedAt: time.Now(),
//		UpdatedAt: time.Now(),
//	}
//
//	if network != "host" && network != "bridge" {
//		app.Mode = string(APP_MODE_FIXED)
//	} else {
//		app.Mode = string(APP_MODE_REPLICATES)
//	}
//
//	if err := s.db.CreateApp(app); err != nil {
//		return nil, err
//	}
//
//	for i := 0; i < count; i++ {
//		res := resource{
//			mem:      version.Mem,
//			cpus:     version.CPUs,
//			disk:     version.Disk,
//			portsNum: len(version.Container.Docker.PortMappings),
//		}
//
//		offer, err := s.RequestOffer(&res)
//		if err != nil {
//			log.Error(err)
//		}
//
//		task, _ := NewTask(
//			version,
//			fmt.Sprintf("%s.%s", i, app.ID),
//			fmt.Sprintf("%s.%s", i, app.ID),
//		)
//
//		task.Build(offer)
//
//		s.addTask(task)
//
//		//if err := s.db.CreateTask(task); err != nil {
//		//	log.Errorf("create task failed: %s", err)
//		//}
//
//		if err := s.LaunchTask(offer, task); err != nil {
//			log.Errorf("launch task got error %s", err)
//		}
//	}
//
//	return app, nil
//}

func (s *Scheduler) InspectApp(appId string) *types.Application {
	return s.db.GetApp(appId)
}

func (s *Scheduler) DeleteApp(appId string) error {
	//app, err := s.db.GetApp(appId)
	//if err != nil {
	//	return err
	//}

	//if err := s.UpdateAppStatus("deleting"); err != nil {
	//	return err
	//}

	//for _, task := range app.tasks {
	//	if err := s.KillTask(task.Id, task.AgentId); err != nil {
	//		return err
	//	}
	//}

	return s.db.DeleteApp(appId)
}

func (s *Scheduler) ListApps(appFilterOptions types.AppFilterOptions) []*types.Application {
	return s.db.ListApps()
}

func (s *Scheduler) ScaleUp(appId string, newInstances int, newIps []string) error {
	return nil
}

func (scheduler *Scheduler) ScaleDown(appId string, removeInstances int) error {
	return nil
}

func (scheduler *Scheduler) UpdateApp(appId string, version *types.Version) error {
	return nil
}

func (scheduler *Scheduler) CancelUpdate(appId string) error {
	return nil
}

func (scheduler *Scheduler) ProceedUpdate(appId string, instances int, newWeights map[string]float64) error {
	return nil
}
