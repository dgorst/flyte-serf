/*
Copyright (C) 2018 Expedia Group.

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

package main

import (
	"log"
	"net/url"
	"os"
	"os/signal"
	api "github.com/HotelsDotCom/flyte-client/client"
	"github.com/HotelsDotCom/flyte-client/flyte"
	"github.com/HotelsDotCom/flyte-serf/agent"
	"github.com/HotelsDotCom/flyte-serf/command"
	"github.com/HotelsDotCom/flyte-serf/event"
	"syscall"
	"time"
)

const packName = "Serf"

func main() {
	handler := &event.UserEventHandler{}

	// Start Serf agent
	a, err := initializeSerfAgent([]agent.HandleEvent{handler.HandleUserEvent})
	if err != nil {
		log.Fatalf("[ERROR] Error initialising Serf agent: %s", err)
	}
	defer a.Shutdown()

	packDef := flyte.PackDef{
		Name:    packName,
		HelpURL: parseURL("https://github.com/HotelsDotCom/flyte-serf/blob/master/README.md"),
		Commands: []flyte.Command{
			command.SendEventCommand(a),
		},
		EventDefs: []flyte.EventDef{
			event.UserEventDef,
		},
	}

	p := flyte.NewPack(packDef, api.NewClient(getHost(), 10*time.Second))
	p.Start()

	handler.SetPack(p)

	// block until we get an exit-causing signal
	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	select {
	case <-signalCh:
	case <-a.ShutdownCh():
	}
	a.Leave()
}

func getHost() *url.URL {
	if os.Getenv("FLYTE_API") == "" {
		log.Fatal("FLYTE_API not set - you must set this to point to an instance of the flyte api server")
	}
	return parseURL(os.Getenv("FLYTE_API"))
}

func parseURL(rawurl string) *url.URL {
	u, err := url.Parse(rawurl)
	if err != nil {
		log.Fatalf("%s is not valid url", rawurl)
	}
	return u
}
