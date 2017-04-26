/*
 *  ZEUS - An Electrifying Build System
 *  Copyright (c) 2017 Philipp Mieden <dreadl0ck [at] protonmail [dot] ch>
 *
 *  This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU General Public License for more details.
 *
 *  You should have received a copy of the GNU General Public License
 *  along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/mgutz/ansi"
)

type parseJobID string

// parseJob represents a parse process for a specific script
// or a command parsed from a zeusfile
// it keeps track of all parsed commands to prevent cycles
// parseJobs can run concurrently
type parseJob struct {

	// path for script to parse, empty if its a command from zesufile
	path string

	// unique identifier
	id parseJobID

	// command array with arguments
	commands [][]string

	// log parse errors to stdout
	silent bool

	// jobs waiting for command currently being parsed by this job
	waiters []chan bool
}

// newJob returns a new parseJob for the given path
func newJob(path string, silent bool) *parseJob {
	return &parseJob{
		path:     path,
		id:       parseJobID(randomString()),
		commands: make([][]string, 0),
		silent:   silent,
	}
}

// AddJob adds a job to the parser
// thread safe
func (p *parser) AddJob(path string, silent bool) (job *parseJob) {

	job = newJob(path, silent)

	Log.WithFields(logrus.Fields{
		"ID":   job.id,
		"path": path,
	}).Debug("adding job")

	p.mutex.Lock()
	p.jobs[job.id] = job
	p.mutex.Unlock()

	return job
}

// RemoveJob removes a job from the parser
// thread safe
func (p *parser) RemoveJob(job *parseJob) {

	Log.WithFields(logrus.Fields{
		"ID":      job.id,
		"path":    job.path,
		"waiters": len(job.waiters),
	}).Debug("removing job")

	// notify waiters
	for _, c := range job.waiters {
		c <- true
	}

	p.mutex.Lock()
	delete(p.jobs, job.id)
	p.mutex.Unlock()
}

func (p *parser) JobExists(path string) bool {

	Log.WithFields(logrus.Fields{
		"path": path,
	}).Debug("job exists?")

	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, job := range p.jobs {
		if job.path == path {
			return true
		}
	}

	return false
}

// wait for a running parseJob
func (p *parser) WaitForJob(path string) {

	p.printJobs()

	c := make(chan bool)

	p.mutex.Lock()

	for _, job := range p.jobs {
		if job.path == path {

			// add channel to waiters
			job.waiters = append(job.waiters, c)

			Log.WithFields(logrus.Fields{
				"ID":   job.id,
				"path": job.path,
			}).Debug("waiting for job")

			p.mutex.Unlock()
			<-c

			Log.WithFields(logrus.Fields{
				"ID":   job.id,
				"path": job.path,
			}).Debug(ansi.Yellow + "job complete" + ansi.Reset)

			return
		}
	}
	p.mutex.Unlock()
}

func (p *parser) printJobs() {

	p.mutex.Lock()
	defer p.mutex.Unlock()

	l.Println(cp.prompt + pad("ID", 20) + " path" + cp.text)
	for _, job := range p.jobs {
		l.Println(pad(string(job.id), 20), job.path)
	}
}
