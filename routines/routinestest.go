package main

import (
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetReportCaller(false)
}

func main() {
	log.Infof("Number of errors is %d", runNestedGoRoutines(2, 3))
}

// Prepares Wait Group and Channel
// Starts Subroutines and waits for execution to stop
func runNestedGoRoutines(numberOfRoutines int, numberOfSubRoutines int) int {

	// Wait Group
	var mainWaitGroup sync.WaitGroup
	mainWaitGroup.Add(numberOfRoutines)

	// Error channel
	errors := make(chan error)

	// Run routines
	for i := 0; i < numberOfRoutines; i++ {
		go runNestedRoutine(i, numberOfSubRoutines, &mainWaitGroup, errors)
	}

	// Wait for execution to end
	go func() {
		log.Info("Started Execution")
		mainWaitGroup.Wait()
		log.Info("Closing Errors Channel")
		close(errors)
		log.Info("Finished")
	}()

	// Return error list
	return len(validateErrorChannel(errors))
}

func runNestedRoutine(identifier int, numberOfSubRoutines int, wg *sync.WaitGroup, errors chan<- error) {

	defer wg.Done()

	log.Infof("[Main %d] Started", identifier)
	invokeGoRoutines(identifier, numberOfSubRoutines, errors)
	invokeSyncronousProcedure(identifier, errors)
	invokeGoRoutinesWithSeparateMethods(numberOfSubRoutines, errors)

	log.Infof("[Main %d] Finished", identifier)
}

// Classical example how to write several goroutines.
// Goroutines share the same waitgroup and error channel
func invokeGoRoutines(identifier int, numberOfSubRoutines int, errors chan<- error) {

	log.Infof("[Inline %d] Procedure Started", identifier)
	var waitGroup sync.WaitGroup
	waitGroup.Add(numberOfSubRoutines)

	for i := 0; i < numberOfSubRoutines; i++ {
		log.Debugf("[Inline %d] Invoking procedure %d", identifier, i)
		go func() {
			defer waitGroup.Done()
			log.Infof("[Inline %d] Procedure %d successful", identifier, i)
			errors <- nil
		}()
	}

	waitGroup.Wait()
	log.Infof("[Inline %d] All subroutines have finished", identifier)
}

func invokeSyncronousProcedure(i int, errors chan<- error) {
	log.Infof("[Sync %d] Simulating synchronous procedure call %d", i, i)
	errors <- nil
}

// Classical example how to write several goroutines which essentially
// invoke methods instead of anonymous functions.
// Goroutines share the same waitgroup and error channel
func invokeGoRoutinesWithSeparateMethods(numberOfSubRoutines int, errors chan<- error) {

	var waitGroup sync.WaitGroup
	waitGroup.Add(numberOfSubRoutines)

	for i := 0; i < numberOfSubRoutines; i++ {
		go invokeSubroutine(i, &waitGroup, errors)
	}

	waitGroup.Wait()
}

func invokeSubroutine(identifier int, wg *sync.WaitGroup, errors chan<- error) {
	defer wg.Done()

	log.Infof("[Delegate %d] Started", identifier)

	if identifier%2 == 0 {
		errors <- fmt.Errorf("Subroutine %d intentional fail on odd number", identifier)
	} else {
		errors <- nil
	}

	log.Debugf("[Delegate %d] Ended", identifier)
}

func validateErrorChannel(errors <-chan error) []error {
	var finalErrors []error

	for {
		nextError, isChannelOpen := <-errors
		if isChannelOpen {
			if nextError != nil {
				log.Errorf("[Error Channel] %s", nextError.Error())
				finalErrors = append(finalErrors, nextError)
			}
		} else {
			break
		}
	}

	log.Info("[Error Channel] Done reading")
	return finalErrors
}
