package core

import (
	"bufio"
	"io"
	"net"
	"strconv"
	"strings"
)

func CreateTools(connection net.Conn, reader *bufio.Reader) ConnectionTools {
	writer := bufio.NewWriter(connection)
	connectionTools := ConnectionTools{Conn: connection, Reader: reader, Writer: writer}
	return connectionTools
}

func ParseFloat(slice []string, position int) (float64, error) {
	result, err := strconv.ParseFloat(strings.TrimSpace(slice[position]), 64)
	if err != nil {
		return 0.0, err
	}
	return result, err
}

func ReadInput(bufferedReader *bufio.Reader) (string, error) {
	readedData, err := bufferedReader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	readedData = strings.TrimSpace(readedData)
	return readedData, err
}

func ReadInputSliced(bufferedReader *bufio.Reader) ([]string, error) {
	readedData, err := bufferedReader.ReadString('\n')
	if err != nil && err != io.EOF {
		return nil, err
	}
	splitedData := strings.Split(readedData, "|")
	if err == io.EOF && splitedData[0] != "" {
		err = nil
	}
	return splitedData, err
}

func ParseInt(slice []string, position int) (int, error) {
	result, err := strconv.ParseInt(strings.TrimSpace(slice[position]), 10, 64)
	if err != nil {
		return 0, err
	}
	return int(result), err
}

func ValidationRocket(rocket Rocket, maximumBurnRate float64) Rocket {
	if rocket.PitchDegree <= 1 {
		rocket.PitchDegree = 1
	}
	if rocket.PitchDegree >= 179 {
		rocket.PitchDegree = 179
	}
	if rocket.YawDegree <= 1 {
		rocket.YawDegree = 1
	}
	if rocket.YawDegree >= 179 {
		rocket.YawDegree = 179
	}
	if rocket.Fuel < 5.0 {
		rocket.Fuel = 5.0
	}
	if rocket.Fuel > rocket.BodyMass*10.0 {
		rocket.Fuel = rocket.BodyMass * 10.0
	}
	if rocket.EngineEfficiency < 500.0 {
		rocket.EngineEfficiency = 500.0
	}
	if rocket.EngineEfficiency > 2500.0 {
		rocket.EngineEfficiency = 2500.0
	}
	if rocket.EngineEfficiency*rocket.BurnRate < (rocket.BodyMass+rocket.Fuel)*GravitationalConstant {
		rocket.BurnRate = (rocket.BodyMass + rocket.Fuel) * GravitationalConstant / rocket.EngineEfficiency
	}
	if rocket.BurnRate > maximumBurnRate {
		rocket.BurnRate = maximumBurnRate
	}

	rocket = Rocket{PitchDegree: rocket.PitchDegree, YawDegree: rocket.YawDegree, Fuel: rocket.Fuel, BurnRate: rocket.BurnRate, BodyMass: rocket.BodyMass, EngineEfficiency: rocket.EngineEfficiency}
	return rocket
}

func ParseRocketCharacteristics(rocketParameters []string) (Rocket, error) {
	pitch, err := ParseFloat(rocketParameters, 0)
	if err != nil {
		return Rocket{}, err
	}

	yaw, err := ParseFloat(rocketParameters, 1)
	if err != nil {
		return Rocket{}, err
	}

	fuel, err := ParseFloat(rocketParameters, 2)
	if err != nil {
		return Rocket{}, err
	}

	burnRate, err := ParseFloat(rocketParameters, 3)
	if err != nil {
		return Rocket{}, err
	}

	bodyMass, err := ParseFloat(rocketParameters, 4)
	if err != nil {
		return Rocket{}, err
	}

	engineEfficiency, err := ParseFloat(rocketParameters, 5)
	if err != nil {
		return Rocket{}, err
	}

	rocket := Rocket{PitchDegree: pitch, YawDegree: yaw, Fuel: fuel, BurnRate: burnRate, BodyMass: bodyMass, EngineEfficiency: engineEfficiency}
	return rocket, err
}

func ParseRocket(rocketParameters []string, maximumBurnRate float64) (Rocket, error) {
	rocket, err := ParseRocketCharacteristics(rocketParameters)
	if err != nil {
		return Rocket{}, err
	}

	rocket = ValidationRocket(rocket, maximumBurnRate)
	return rocket, nil
}

func ParseNRocket(totalRocket int, connectionReader *bufio.Reader) ([]Rocket, error) {
	rockets := make([]Rocket, totalRocket)

	for countRocket := 0; countRocket < totalRocket; countRocket++ {

		rocketParameters, err := ReadInputSliced(connectionReader)
		if err != nil {
			return rockets, err
		}
		rocket, err := ParseRocketCharacteristics(rocketParameters)
		if err != nil {
			return rockets, err
		}
		rockets[countRocket] = rocket
	}

	return rockets, nil
}

func ParseNRocketValidated(totalRocket int, connectionReader *bufio.Reader, maximumBurnRate float64) ([]Rocket, error) {
	rockets := make([]Rocket, totalRocket)

	for countRocket := 0; countRocket < totalRocket; countRocket++ {

		rocketParameters, err := ReadInputSliced(connectionReader)
		if err != nil {
			return rockets, err
		}
		rocket, err := ParseRocket(rocketParameters, maximumBurnRate)
		if err != nil {
			return rockets, err
		}
		rockets[countRocket] = rocket
	}

	return rockets, nil
}

func ReadInputCommaSliced(bufferedReader *bufio.Reader) ([]string, error) {
	readedData, err := bufferedReader.ReadString('\n')
	if err != nil && err != io.EOF {
		return nil, err
	}
	splitedData := strings.Split(readedData, ",")
	if err == io.EOF && splitedData[0] != "" {
		err = nil
	}
	return splitedData, err
}
