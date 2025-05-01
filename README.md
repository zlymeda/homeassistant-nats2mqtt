# homeassistant-nats2mqtt

A Golang library that bridges [NATS](https://nats.io/) messaging with Home Assistant MQTT 
integration, enabling 
seamless device and entity management through auto-discovery.

The main goal is not to pollute Home Assistant with more and more integrations. Instead, this library enables you to run separate services written in Go that can operate independently from Home Assistant, be restarted without affecting HA, and don't have to be written in Python.

## Overview

This library allows you to define Home Assistant entities in Go code and automatically publish them to Home Assistant using [NATS](https://nats.io/) as the underlying transport mechanism. It handles the MQTT discovery protocol, state monitoring, and command processing, making it easy to integrate custom devices with Home Assistant.

## Features

- Auto-discovery - Automatically register your devices and entities with Home Assistant
- State monitoring - Track and publish entity state changes
- Command processing - Listen for commands from Home Assistant and execute callbacks
- Multiple entity types supported:
  - Alarm
  - Binary sensor
  - Button
  - Climate
  - Cover
  - Device tracker
  - Lock
  - Number
  - Sensor
  - Switch

(not all features are supported as of now)

## Installation

```shell
go get github.com/zlymeda/homeassistant-nats2mqtt
```

## Architecture

The library consists of several key components:

1. Entity definitions - Go structs that define different types of Home Assistant entities
1. Observable pattern - For tracking state changes
1. Service executor - Manages [NATS](https://nats.io/) connections and handles message publishing/subscribing
1. Entity registry - Tracks entities for a device and manages their lifecycle

## Usage
### Basic Setup

```golang
package main

import (
	"context"
	"github.com/nats-io/nats.go"
	"github.com/zlymeda/homeassistant-nats2mqtt/entity"
	"github.com/zlymeda/homeassistant-nats2mqtt/executor"
	"log"
)

func main() {
	ctx := context.Background()

	// Connect to NATS
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}

	// Create service
	exec := executor.New(ctx, executor.Origin{
		Name:            "myapp2mqtt",
		SoftwareVersion: "dev",
	}, nc)

	// Register devices and entities
	// ...
  
	exec.Start()
}

```

### Creating a Device with Entities

```golang
// Create a device
car := exec.AddDevice(entity.Device{
      Id:              "vin123456",
      Name:            "car_name",
      DisplayName:     "Car Name",
      Manufacturer:    "XYZ",
	  // ...
 })

// Add entity to the device
err = car.AddCover(NewChargerDoor(ctx))
if err != nil {
    log.Fatalf("Failed to add cover: %v", err)
}


```

### Example: Creating a Cover Entity
```golang
func NewChargerDoor(ctx context.Context) *entity.Cover {
	state := observable.NewSimple[entity.CoverState](ctx, entity.CoverStateClosed)
	
	return &entity.Cover{
		Meta: observable.NewSingle[entity.Meta](entity.Meta{
			Id:          "charger_door",
			Name:        "Charger door",
			DeviceClass: "door",
			Icon:        "mdi:ev-plug-ccs1",
		}),
		State: state,
		Open: func() error {
			state.Change(entity.CoverStateOpen)
			return nil // Actual implementation would control the physical device
		},
		Close: func() error {
			state.Change(entity.CoverStateClosed)
			return nil // Actual implementation would control the physical device
		},
	}
}

```

### Real-world Example
```golang
func (c *Car) NewChargerDoor(ctx context.Context) *entity.Cover {
	state := observable.NewSimple[entity.CoverState](ctx, "none")

	c.addBooleanHandler("ChargePortDoorOpen", func(open bool) {
		if open {
			state.Change(entity.CoverStateOpen)
		} else {
			state.Change(entity.CoverStateClosed)
		}
	})

	return &entity.Cover{
		Meta: observable.NewSingle[entity.Meta](entity.Meta{
			Id:          "charger_door",
			Name:        "Charger door",
			DeviceClass: "door",
			Icon:        "mdi:ev-plug-ccs1",
		}),
		State: state,
		Open: func() error {
			current := state.Current()
			if current == entity.CoverStateOpen {
				slog.Info("charge port already opened")
				return nil
			}
			return c.Execute(command.ChargePortOpen())
		},
		Close: func() error {
			current := state.Current()
			if current == entity.CoverStateClosed {
				slog.Info("charge port already closed")
				return nil
			}
			return c.Execute(command.ChargePortClose())
		},
	}
}

```

### Observable

The state, meta and attributes use observables.

The state will usually use `observable.NewSimple` as the state changes.
The metadata will most often be just a `observable.NewSingle`, meaning it won't change. 
However if you want to change an icon then use `observable.NewSimple` and swap the icon, example:

```golang
meta := observable.NewSimple[entity.Meta]

// call whenever soc/charging changes, pick latest meta and update the icon based on the state and change it
// this will republish the autodiscovery and updates the icon in HA
updateMeta := func() {
	metadata := meta.Current()
	metadata.Icon = "mdi:" + pickIcon(soc, charging)
	meta.Change(metadata)
}
```


## How It Works
1. The library creates entities with observable state
1. When you register an entity, it:
   - Publishes discovery information to Home Assistant
   - Sets up monitoring for state changes
   - Creates command handlers for interactive entities
1. Home Assistant discovers the entities and creates the appropriate UI elements
1. State changes in your Go code are automatically published to Home Assistant
1. Commands from Home Assistant are routed to your callback functions

## NATS and MQTT

This library uses [NATS](https://nats.io/) as the messaging system but publishes messages in a format compatible 
with Home Assistant's MQTT integration. It is compatible with existing Home Assistant MQTT 
discovery.

## License
MIT

## Contributing
Contributions are welcome! Please feel free to submit a Pull Request.

