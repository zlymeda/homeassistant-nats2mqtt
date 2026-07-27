package entity

import (
	"testing"
)

func TestMeta_GetId(t *testing.T) {
	meta := Meta{Id: "test-id"}
	if got := meta.GetId(); got != "test-id" {
		t.Errorf("GetId() = %v, want %v", got, "test-id")
	}
}

func TestMeta_GetName(t *testing.T) {
	meta := Meta{Name: "Test Name"}
	if got := meta.GetName(); got != "Test Name" {
		t.Errorf("GetName() = %v, want %v", got, "Test Name")
	}
}

func TestMeta_GetExtraTopics(t *testing.T) {
	meta := Meta{}
	if got := meta.GetExtraTopics(); got != nil {
		t.Errorf("GetExtraTopics() = %v, want nil", got)
	}
}

func TestMeta_IsMqttOnly(t *testing.T) {
	tests := []struct {
		name     string
		mqttOnly bool
		want     bool
	}{
		{"mqtt only true", true, true},
		{"mqtt only false", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := Meta{MqttOnly: tt.mqttOnly}
			if got := meta.IsMqttOnly(); got != tt.want {
				t.Errorf("IsMqttOnly() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMeta_ToHaDiscovery(t *testing.T) {
	device := Device{Id: "dev-1", Name: "Test Device"}

	tests := []struct {
		name     string
		meta     Meta
		platform string
		check    func(map[string]any) bool
	}{
		{
			name: "basic fields",
			meta: Meta{
				Id:       "sensor1",
				Name:     "Sensor 1",
				ObjectId: "obj1",
				UniqueId: "uid1",
			},
			platform: "sensor",
			check: func(m map[string]any) bool {
				return m["p"] == "sensor" && m["name"] == "Sensor 1" &&
					m["obj_id"] == "obj1" && m["uniq_id"] == "uid1"
			},
		},
		{
			name: "with device class",
			meta: Meta{
				Id:          "temp",
				Name:        "Temperature",
				DeviceClass: "temperature",
			},
			platform: "sensor",
			check: func(m map[string]any) bool {
				return string(m["dev_cla"].(DeviceClass)) == "temperature"
			},
		},
		{
			name: "with icon",
			meta: Meta{
				Id:   "switch1",
				Name: "Switch 1",
				Icon: "mdi:toggle-switch",
			},
			platform: "switch",
			check: func(m map[string]any) bool {
				return m["ic"] == "mdi:toggle-switch"
			},
		},
		{
			name: "with unit",
			meta: Meta{
				Id:   "power",
				Name: "Power",
				Unit: "W",
			},
			platform: "sensor",
			check: func(m map[string]any) bool {
				return m["unit_of_meas"] == "W"
			},
		},
		{
			name: "with state class",
			meta: Meta{
				Id:         "energy",
				Name:       "Energy",
				StateClass: "total_increasing",
			},
			platform: "sensor",
			check: func(m map[string]any) bool {
				return string(m["stat_cla"].(StateClass)) == "total_increasing"
			},
		},
		{
			name: "with entity category",
			meta: Meta{
				Id:             "config1",
				Name:           "Config 1",
				EntityCategory: CategoryConfig,
			},
			platform: "number",
			check: func(m map[string]any) bool {
				return m["ent_cat"] == CategoryConfig
			},
		},
		{
			name: "auto-generated object id and unique id",
			meta: Meta{
				Id:   "sensor2",
				Name: "Sensor 2",
			},
			platform: "sensor",
			check: func(m map[string]any) bool {
				return m["obj_id"] == "Test Device_sensor2" &&
					m["uniq_id"] == "dev-1_sensor2"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.meta.ToHaDiscovery(device, tt.platform)
			if !tt.check(result) {
				t.Errorf("ToHaDiscovery() check failed, got %v", result)
			}
		})
	}
}

func TestNumberMeta_GetId(t *testing.T) {
	meta := NumberMeta{Meta: Meta{Id: "num-1"}}
	if got := meta.GetId(); got != "num-1" {
		t.Errorf("GetId() = %v, want %v", got, "num-1")
	}
}

func TestNumberMeta_GetName(t *testing.T) {
	meta := NumberMeta{Meta: Meta{Name: "Number 1"}}
	if got := meta.GetName(); got != "Number 1" {
		t.Errorf("GetName() = %v, want %v", got, "Number 1")
	}
}

func TestNumberMeta_GetExtraTopics(t *testing.T) {
	meta := NumberMeta{ExtraTopics: ExtraTopicsList{"cmd"}}
	if got := meta.GetExtraTopics(); len(got) != 1 || got[0] != "cmd" {
		t.Errorf("GetExtraTopics() = %v, want [cmd]", got)
	}
}

func TestSelectMeta_GetId(t *testing.T) {
	meta := SelectMeta{Meta: Meta{Id: "sel-1"}}
	if got := meta.GetId(); got != "sel-1" {
		t.Errorf("GetId() = %v, want %v", got, "sel-1")
	}
}

func TestSelectMeta_GetName(t *testing.T) {
	meta := SelectMeta{Meta: Meta{Name: "Select 1"}}
	if got := meta.GetName(); got != "Select 1" {
		t.Errorf("GetName() = %v, want %v", got, "Select 1")
	}
}

func TestSelectMeta_GetExtraTopics(t *testing.T) {
	meta := SelectMeta{}
	if got := meta.GetExtraTopics(); got != nil {
		t.Errorf("GetExtraTopics() = %v, want nil", got)
	}
}

func TestSwitchMeta_GetId(t *testing.T) {
	meta := SwitchMeta{Meta: Meta{Id: "sw-1"}}
	if got := meta.GetId(); got != "sw-1" {
		t.Errorf("GetId() = %v, want %v", got, "sw-1")
	}
}

func TestSwitchMeta_GetName(t *testing.T) {
	meta := SwitchMeta{Meta: Meta{Name: "Switch 1"}}
	if got := meta.GetName(); got != "Switch 1" {
		t.Errorf("GetName() = %v, want %v", got, "Switch 1")
	}
}

func TestSwitchMeta_GetExtraTopics(t *testing.T) {
	meta := SwitchMeta{}
	if got := meta.GetExtraTopics(); len(got) != 1 || got[0] != CommandTopic {
		t.Errorf("GetExtraTopics() = %v, want [%s]", got, CommandTopic)
	}
}
