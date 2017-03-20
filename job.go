/*
 *  ZEUS - An Electrifying Build System
 *  Copyright (c) 2017 Philipp Mieden <dreadl0ck@protonmail.ch>
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

// parseJob represents a parse process for a specific script
// it keeps track of all parsed commands to prevent cycles
// parseJobs can run concurrently
type parseJob struct {

	// path for script to parse
	path string

	// command array with arguments
	commands [][]string

	// log parse errors to stdout
	silent bool
}

// newJob returns a new parseJob for the given path
func newJob(path string, silent bool) *parseJob {
	return &parseJob{
		path:     path,
		commands: make([][]string, 0),
		silent:   silent,
	}
}

// AddJob adds a job to the parser
// thread safe
func (p *parser) AddJob(path string, silent bool) (job *parseJob) {

	job = newJob(path, silent)

	Log.Debug("adding job: ", job.path)

	p.mutex.Lock()
	p.jobs[path] = job
	p.mutex.Unlock()

	return job
}

// RemoveJob removes a job from the parser
// thread safe
func (p *parser) RemoveJob(job *parseJob) {

	Log.Debug("removing job: ", job.path)

	p.mutex.Lock()
	delete(p.jobs, job.path)
	p.mutex.Unlock()
}
