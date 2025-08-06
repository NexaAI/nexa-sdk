package nexa_sdk

import "testing"

func TestGetPluginList(t *testing.T) {
	plugins, err := GetPluginList()
	if err != nil {
		t.Fatalf("GetPluginList failed: %v", err)
	}

	t.Logf("Available plugins: %v", plugins.PluginIDs)
}

func TestGetDeviceList(t *testing.T) {
	// First get the plugin list
	plugins, err := GetPluginList()
	if err != nil {
		t.Fatalf("GetPluginList failed: %v", err)
	}

	if len(plugins.PluginIDs) == 0 {
		t.Skip("No plugins available, skipping device list test")
	}

	for _, pluginID := range plugins.PluginIDs {
		devices, err := GetDeviceList(DeviceListInput{PluginID: pluginID})
		if err != nil {
			t.Fatalf("GetDeviceList failed for plugin %s: %v", pluginID, err)
		}

		t.Logf("Plugin %s devices:", pluginID)
		for _, device := range devices.Devices {
			t.Logf("  Device ID: %s, Name: %s", device.ID, device.Name)
		}
	}
}
