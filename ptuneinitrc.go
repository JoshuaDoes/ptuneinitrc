package main

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	VERSION = "1.0"
	PTUNE = `  task_profiles ProcessCapacityMax MaxPerformance
  ioprio rt 4
  rlimit rtprio 10 10
  rlimit memlock unlimited unlimited
  priority -20
  writepid /dev/cpuset/system/tasks`
	HEADER = `#
# ptune.rc by JoshuaDoes for use with Pixel Tune
# Careful with your changes!
#
# Generated using ptuneinitrc v%s
# Built for %s
# Unix epoch in milliseconds: %d
#
# Current template:
%s
#

#
## INIT STAGES
#

on early-init
  # Initialize our system cpuset
  mkdir /dev/cpuset/system
  chown system system /dev/cpuset/system/tasks
  chmod 0664 /dev/cpuset/system/tasks
  write /dev/cpuset/system/cpus %s
  write /dev/cpuset/system/mems 0

#
## Specific to Android %s (%s)
#`
)

func main() {
	initrcs := findInitRCs()
	if len(initrcs) == 0 {
		return
	}
	ignore := make([]string, 0)

	//Handle known variable imports
	zygote := getprop("ro.zygote")
	if zygote != "" {
		for i := 0; i < len(initrcs); i++ {
			if name := initrcs[i].name; strings.HasPrefix(name, "init.zygote") {
				if name != fmt.Sprintf("init.%s.rc", zygote) {
					ignore = append(ignore, name)
				}
			}
		}
	}

	//Remove ignored initrcs list
	filtered := make([]*initrc, 0)
	for i := 0; i < len(initrcs); i++ {
		skip := false
		for j := 0; j < len(ignore); j++ {
			if initrcs[i].name == ignore[j] {
				skip = true
				break
			}
		}
		if !skip {
			filtered = append(filtered, initrcs[i])
		} else {
			initrcs[i].close()
		}
	}
	initrcs = filtered

	svcLock := sync.Mutex{}
	services := make([]*service, 0)
	jobs := len(initrcs)
	for _, rc := range initrcs {
		go func(rc *initrc, lock *sync.Mutex) {
			svcs := rc.getServices()
			lock.Lock()
			services = append(services, svcs...)
			lock.Unlock()
			rc.close()
			jobs--
		}(rc, &svcLock)
	}
	for jobs > 0 {
		time.Sleep(time.Millisecond * 1)
	}
	if len(services) == 0 {
		return
	}
	sort.SliceStable(services, func(i, j int) bool {
		return services[i].name < services[j].name
	})

	fingerprint := getprop("ro.build.fingerprint")
	version := getprop("ro.build.version.release")
	if codename := getprop("ro.build.version.release_or_codename"); version != codename {
		version = fmt.Sprintf("%s/%s", version, codename)
	}
	buildid := getprop("ro.build.id")
	cpus := fmt.Sprintf("0-%d", runtime.NumCPU()-1)

	template := strings.Split(PTUNE, "\n")
	for i := 0; i < len(template); i++ {
		template[i] = "# " + template[i]
	}

	fmt.Printf(HEADER, VERSION, fingerprint, time.Now().UnixMilli(), strings.Join(template, "\n"), cpus, version, buildid)
	fmt.Printf("\n\n")

	for i := 0; i < len(services); i++ {
		svc := services[i]
		fmt.Printf("# %s\n%s\n", svc.path, svc.line)
		for j := 0; j < len(svc.block); j++ {
			fmt.Printf("  %s\n", svc.block[j])
		}
		fmt.Printf("%s", PTUNE)
		fmt.Printf("\n\n")
	}
}
