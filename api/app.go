package api

import (
	//"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Dataman-Cloud/swan/mesos"
	"github.com/Dataman-Cloud/swan/types"
	//"github.com/Dataman-Cloud/swan/utils/fields"
	//"github.com/Dataman-Cloud/swan/utils/labels"

	log "github.com/Sirupsen/logrus"
)

func (r *Router) createApp(w http.ResponseWriter, req *http.Request) error {
	if err := r.CheckForJSON(req); err != nil {
		log.Errorf("Check http header failed %v", err)
		return r.WriteError(http.StatusInternalServerError, err)
	}

	if err := req.ParseForm(); err != nil {
		log.Errorf("Parse request failed: %v", err)
		return r.WriteError(http.StatusInternalServerError, err)
	}

	name := req.Form.Get("name")

	if name == "" {
		name = "default"
	}

	var version types.Version
	if err := r.decode(req.Body, &version); err != nil {
		log.Errorf("Decoding json failed: %v", err)
		return r.WriteError(http.StatusInternalServerError, err)
	}

	var (
		cfg     = &version
		id      = fmt.Sprintf("%s.%s.%s.%s", cfg.AppName, name, cfg.RunAs, r.driver.ClusterName())
		network = strings.ToLower(cfg.Container.Docker.Network)
		count   = int(cfg.Instances)
	)

	if app := r.db.GetApp(id); app != nil {
		err := fmt.Errorf("app %s has already exists", id)
		log.Errorf("%v", err)
		return r.WriteError(http.StatusConflict, err)
	}

	mode := "replicas"
	if network != "host" && network != "bridge" {
		mode = "fixed"
	}

	app := &types.Application{
		ID:        id,
		Name:      cfg.AppName,
		Version:   cfg,
		RunAs:     cfg.RunAs,
		ClusterID: r.driver.ClusterName(),
		Status:    "creating",
		Mode:      mode,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := r.db.CreateApp(app); err != nil {
		log.Errorf("create app error: %v", err)
		return r.WriteError(http.StatusInternalServerError, err)
	}

	go func() {
		for i := 0; i < count; i++ {
			var (
				id   = fmt.Sprintf("%d.%s", i, app.ID)
				name = fmt.Sprintf("%d.%s", i, app.ID)
			)

			task := mesos.NewTask(
				cfg,
				id,
				name,
			)
			if err := r.driver.LaunchTask(task); err != nil {
				log.Errorf("launch task %s got error: %v", id, err)
				return
			}

			if err := r.db.CreateTask(&types.Task{
				ID:      id,
				Name:    name,
				CPU:     cfg.CPUs,
				Mem:     cfg.Mem,
				Disk:    cfg.Disk,
				Image:   cfg.Container.Docker.Image,
				Weight:  100,
				Created: time.Now(),
			}); err != nil {
				log.Errorf("create task failed: %s", err)
			}
		}
	}()

	return r.WriteJSON(w, http.StatusCreated, app)
}

func (r *Router) listApps(w http.ResponseWriter, req *http.Request) error {
	return nil
}

func (r *Router) getApp(w http.ResponseWriter, req *http.Request) error {
	return nil
}

func (r *Router) deleteApp(w http.ResponseWriter, req *http.Request) error {
	return nil
}

func (r *Router) scaleUp(w http.ResponseWriter, req *http.Request) error {
	var param types.ScaleUpParam

	if err := r.decode(req.Body, &param); err != nil {
		return err
	}

	var (
		count = &param.Instances
		ips   = &param.IPs
	)

	log.Info(count, ips)

	return nil
}

func (r *Router) scaleDown(w http.ResponseWriter, req *http.Request) error {
	return nil
}

func (r *Router) updateApp(w http.ResponseWriter, req *http.Request) error {
	return nil
}

func (r *Router) proceedUpdate(w http.ResponseWriter, req *http.Request) error {
	return nil
}

func (r *Router) cancelUpdate(w http.ResponseWriter, req *http.Request) error {
	return nil
}
func (r *Router) updateWeights(w http.ResponseWriter, req *http.Request) error {
	return nil
}
func (r *Router) getTask(w http.ResponseWriter, req *http.Request) error {
	return nil
}
func (r *Router) updateWeight(w http.ResponseWriter, req *http.Request) error {
	return nil
}
func (r *Router) getVersions(w http.ResponseWriter, req *http.Request) error {
	return nil
}
func (r *Router) getAppVersion(w http.ResponseWriter, req *http.Request) error {
	return nil
}
