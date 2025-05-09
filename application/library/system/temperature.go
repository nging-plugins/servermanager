package system

import "github.com/shirou/gopsutil/v4/sensors"

func SensorsTemperatures() ([]sensors.TemperatureStat, error) {
	r := []sensors.TemperatureStat{}
	temps, err := sensors.SensorsTemperatures()
	if err != nil {
		return r, err
	}
	for _, temp := range temps {
		if temp.Temperature == 0 {
			continue
		}
		temp.SensorKey = keyFormatter(temp.SensorKey)
		r = append(r, temp)
	}
	return r, nil
}
