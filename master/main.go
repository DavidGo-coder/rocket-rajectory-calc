package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"net"
	"strings"
	"time"

	core "github.com/DavidGo-coder/rocket-trajecory-calc"
)

const (
	portConnection    string = "8800"
	timeOneGeneration        = 3 * time.Second
	medianHisorySize  int    = 5
	totalProgram             = 5

	reset      = "\x1b[0m"
	red        = "\x1b[31m"
	colorGreen = "\x1b[32m"
	yellow     = "\x1b[33m"
	blue       = "\x1b[34m"
	purple     = "\x1b[35m"
	cyan       = "\x1b[36m"
	bold       = "\x1b[1m"
)

func parseAdvancedRocketCharaceristics(rocketParameters []string) (core.Rocket, error) {
	pitch, err := core.ParseFloat(rocketParameters, 0)
	if err != nil {
		return core.Rocket{}, err
	}

	yaw, err := core.ParseFloat(rocketParameters, 1)
	if err != nil {
		return core.Rocket{}, err
	}

	fuel, err := core.ParseFloat(rocketParameters, 2)
	if err != nil {
		return core.Rocket{}, err
	}

	burnRate, err := core.ParseFloat(rocketParameters, 3)
	if err != nil {
		return core.Rocket{}, err
	}

	bodyMass, err := core.ParseFloat(rocketParameters, 4)
	if err != nil {
		return core.Rocket{}, err
	}

	engineEfficiency, err := core.ParseFloat(rocketParameters, 5)
	if err != nil {
		return core.Rocket{}, err
	}

	x, err := core.ParseFloat(rocketParameters, 6)
	if err != nil {
		return core.Rocket{}, err
	}

	y, err := core.ParseFloat(rocketParameters, 7)
	if err != nil {
		return core.Rocket{}, err
	}

	z, err := core.ParseFloat(rocketParameters, 8)
	if err != nil {
		return core.Rocket{}, err
	}

	fitness, err := core.ParseFloat(rocketParameters, 9)
	if err != nil {
		return core.Rocket{}, err
	}

	trajectoryCSV := rocketParameters[10]

	rocket := core.Rocket{PitchDegree: pitch, YawDegree: yaw, Fuel: fuel, BurnRate: burnRate, BodyMass: bodyMass, EngineEfficiency: engineEfficiency, X: x, Y: y, Z: z, Fitness: fitness, TrajectoryCSV: trajectoryCSV}
	return rocket, err
}

func parseNAdvancedRocket(totalRocket int, connectionReader *bufio.Reader) ([]core.Rocket, error) {
	rockets := make([]core.Rocket, totalRocket)

	for countRocket := 0; countRocket < totalRocket; countRocket++ {

		rocketParameters, err := core.ReadInputSliced(connectionReader)
		if err != nil {
			return rockets, err
		}
		rocket, err := parseAdvancedRocketCharaceristics(rocketParameters)
		if err != nil {
			return rockets, err
		}
		rockets[countRocket] = rocket
	}

	return rockets, nil
}

func finishWorking(writer1, writer2, writer3, writer4 *bufio.Writer) error {
	_, err := fmt.Fprintln(writer1, "FINISHWORKING")
	if err != nil {
		return err
	}
	writer1.Flush()

	_, err = fmt.Fprintln(writer2, "FINISHWORKING")
	if err != nil {
		return err
	}
	writer2.Flush()

	_, err = fmt.Fprintln(writer3, "FINISHWORKING")
	if err != nil {
		return err
	}
	writer3.Flush()

	_, err = fmt.Fprintln(writer4, "FINISHWORKING")
	if err != nil {
		return err
	}
	writer4.Flush()
	return nil
}

func waitSlowPython(bufferedReader *bufio.Reader) error {
	currentTime := time.Now()

	_, err := bufferedReader.ReadString('\n')
	if err != nil && err != io.EOF {
		fmt.Printf("🤬 %s[PYTHON CRASHED]%s Oh come on! The visualization client just walked out on us: %v\n", red, bold, err)
		return err
	}

	if time.Since(currentTime) < timeOneGeneration {
		timeSpent := time.Since(currentTime)
		fmt.Printf("🥱 %s[WAITING FOR PYTHON]%s Calculations are too fast, Python is lagging behind... Napping for %v\n", yellow, reset, timeOneGeneration-timeSpent)
		time.Sleep(timeOneGeneration - timeSpent)
	}
	fmt.Printf("✨ %s[SYNC SUCCESS]%s Python finally caught up! Generation data rendered beautifully! 🎉\n", colorGreen, reset)
	return err
}

func parseMedianLength(bufferedReader *bufio.Reader) (float64, int, error) {
	parameters, err := core.ReadInputSliced(bufferedReader)
	if err != nil {
		return 0, 0, err
	}

	median, err := core.ParseFloat(parameters, 0)
	if err != nil {
		return 0, 0, err
	}

	length, err := core.ParseInt(parameters, 1)
	if err != nil {
		return 0, 0, err
	}

	return median, length, nil
}

func startCalculation(bufferedReader *bufio.Reader, bufferedWriter *bufio.Writer, targetX, targetY, targetZ, populationSize int, rocketPopulation []core.Rocket) (float64, []core.Rocket, error) {
	_, err := fmt.Fprintln(bufferedWriter, "COMMAND1")
	if err != nil {
		return 0.0, []core.Rocket{}, err
	}
	bufferedWriter.Flush()

	_, err = fmt.Fprintf(bufferedWriter, "%v|%v|%v|%v|%v\n", targetX, targetY, targetZ, populationSize, len(rocketPopulation))
	if err != nil {
		return 0.0, []core.Rocket{}, err
	}
	bufferedWriter.Flush()

	for i := 0; i < len(rocketPopulation); i++ {
		_, err = fmt.Fprintf(bufferedWriter, "%v|%v|%v|%v|%v|%v\n", rocketPopulation[i].PitchDegree, rocketPopulation[i].YawDegree, rocketPopulation[i].Fuel, rocketPopulation[i].BurnRate, rocketPopulation[i].BodyMass, rocketPopulation[i].EngineEfficiency)
		if err != nil {
			return 0.0, []core.Rocket{}, err
		}
		bufferedWriter.Flush()
	}

	median, length, err := parseMedianLength(bufferedReader)
	if err != nil {
		return 0.0, []core.Rocket{}, err
	}

	bestRockets, err := parseNAdvancedRocket(length, bufferedReader)
	if err != nil {
		return 0.0, []core.Rocket{}, err
	}

	return median, bestRockets, nil
}

func startFirstAlgorithm(bufferedReader *bufio.Reader, bufferedWriter *bufio.Writer, firstRocket core.Rocket, populationSize int, maximumBurnRate float64) ([]core.Rocket, error) {
	_, err := fmt.Fprintln(bufferedWriter, "COMMAND1")
	if err != nil {
		return []core.Rocket{}, err
	}
	bufferedWriter.Flush()

	_, err = fmt.Fprintf(bufferedWriter, "%v|%v\n", populationSize, maximumBurnRate)
	if err != nil {
		return []core.Rocket{}, err
	}
	bufferedWriter.Flush()

	_, err = fmt.Fprintf(bufferedWriter, "%v|%v|%v|%v|%v|%v\n", firstRocket.PitchDegree, firstRocket.YawDegree, firstRocket.Fuel, firstRocket.BurnRate, firstRocket.BodyMass, firstRocket.EngineEfficiency)
	if err != nil {
		return []core.Rocket{}, err
	}
	bufferedWriter.Flush()

	rocketPopulation, err := core.ParseNRocket(populationSize, bufferedReader)
	if err != nil {
		return []core.Rocket{}, err
	}

	return rocketPopulation, err
}

func startAlgorithm(bufferedReader *bufio.Reader, bufferedWriter *bufio.Writer, bestRockets []core.Rocket, populationSize, countGenerations int, medianHistory []float64) ([]core.Rocket, error) {
	_, err := fmt.Fprintln(bufferedWriter, "COMMAND2")
	if err != nil {
		return []core.Rocket{}, err
	}
	bufferedWriter.Flush()

	_, err = fmt.Fprintf(bufferedWriter, "%v|%v\n", countGenerations, len(bestRockets))
	if err != nil {
		return []core.Rocket{}, err
	}
	bufferedWriter.Flush()

	for i := 0; i < len(medianHistory); i++ {
		if i != 0 {
			_, err = fmt.Fprint(bufferedWriter, "|")
			if err != nil {
				return []core.Rocket{}, err
			}
			bufferedWriter.Flush()
		}
		if i+1 == len(medianHistory) {
			_, err = fmt.Fprintf(bufferedWriter, "%v\n", medianHistory[i])
			if err != nil {
				return []core.Rocket{}, err
			}
			bufferedWriter.Flush()
			break
		}

		_, err = fmt.Fprintf(bufferedWriter, "%v", medianHistory[i])
		if err != nil {
			return []core.Rocket{}, err
		}
		bufferedWriter.Flush()
	}

	for i := 0; i < len(bestRockets); i++ {
		_, err = fmt.Fprintf(bufferedWriter, "%v|%v|%v|%v|%v|%v\n", bestRockets[i].PitchDegree, bestRockets[i].YawDegree, bestRockets[i].Fuel, bestRockets[i].BurnRate, bestRockets[i].BodyMass, bestRockets[i].EngineEfficiency)
		if err != nil {
			return []core.Rocket{}, err
		}
		bufferedWriter.Flush()
	}

	rocketPopulation, err := core.ParseNRocket(populationSize, bufferedReader)
	if err != nil {
		return []core.Rocket{}, err
	}
	bufferedWriter.Flush()

	return rocketPopulation, err
}

func readAllInput(bufferedReader *bufio.Reader) (targetX, targetY, targetZ, populationSize, generations int, firstRocket core.Rocket, maximumBurnRate float64, err error) {
	targetCoordinates, err := core.ReadInputSliced(bufferedReader)
	if err != nil {
		return targetX, targetY, targetZ, populationSize, generations, firstRocket, maximumBurnRate, err
	}

	targetX, err = core.ParseInt(targetCoordinates, 0)
	if err != nil {
		return targetX, targetY, targetZ, populationSize, generations, firstRocket, maximumBurnRate, err
	}

	targetY, err = core.ParseInt(targetCoordinates, 1)
	if err != nil {
		return targetX, targetY, targetZ, populationSize, generations, firstRocket, maximumBurnRate, err
	}

	targetZ, err = core.ParseInt(targetCoordinates, 2)
	if err != nil {
		return targetX, targetY, targetZ, populationSize, generations, firstRocket, maximumBurnRate, err
	}

	rocketParameters, err := core.ReadInputSliced(bufferedReader)
	if err != nil {
		return targetX, targetY, targetZ, populationSize, generations, firstRocket, maximumBurnRate, err
	}

	distance := math.Hypot(float64(targetX), float64(targetY))
	maximumBurnRate = 20.0 + (distance / 50.0)

	firstRocket, err = core.ParseRocket(rocketParameters, maximumBurnRate)
	if err != nil {
		return targetX, targetY, targetZ, populationSize, generations, firstRocket, maximumBurnRate, err
	}

	generalParameters, err := core.ReadInputSliced(bufferedReader)
	if err != nil {
		return targetX, targetY, targetZ, populationSize, generations, firstRocket, maximumBurnRate, err
	}

	populationSize, err = core.ParseInt(generalParameters, 0)
	if err != nil {
		return targetX, targetY, targetZ, populationSize, generations, firstRocket, maximumBurnRate, err
	}

	generations, err = core.ParseInt(generalParameters, 1)
	if err != nil {
		return targetX, targetY, targetZ, populationSize, generations, firstRocket, maximumBurnRate, err
	}

	return targetX, targetY, targetZ, populationSize, generations, firstRocket, maximumBurnRate, nil
}

func main() {
	var geneticTools core.ConnectionTools
	var inputTools core.ConnectionTools
	var simulationTools core.ConnectionTools
	var visualizationTools core.ConnectionTools
	var printerTools core.ConnectionTools

	listener, err := net.Listen("tcp", ":"+portConnection)
	if err != nil {
		fmt.Printf("🤬 %s[LAUNCH CRASH]%s Can't bind port %v! Did someone steal it?! Error: %v\n", red, bold, portConnection, err)
		return
	}

	fmt.Printf("📡 %s[I AM ALIVE!]%s TCP Server successfully bound to port %s%v%s. Awaiting guests...\n", cyan, reset, purple, portConnection, reset)
	defer listener.Close()

	for countProgram := 0; countProgram < totalProgram; countProgram++ {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Printf("😭 %s[HANDSHAKE ABORTED]%s A module tried to link up but crashed down: %v\n", red, reset, err)
		}
		connectionReader := bufio.NewReader(connection)

		obtainedData, err := connectionReader.ReadString('\n')
		if err != nil {
			fmt.Printf("🧐 %s[WHAT'S THAT NOISE?]%s Received totally unreadable garbage bytes: %v\n", yellow, reset, err)
			return
		}
		obtainedData = strings.TrimSpace(obtainedData)

		switch obtainedData {
		case "I|AM|PARSER":
			inputTools = core.CreateTools(connection, connectionReader)
			fmt.Printf("📥 %s[HELLO, PARSER!]%s Input module has entered the lobby. Config specs received!\n", blue, reset)
		case "I|AM|ALGORITHM":
			geneticTools = core.CreateTools(connection, connectionReader)
			fmt.Printf("🧬 %s[EVOLUTION IS HERE]%s Genetic algorithm hooked up. Ready to mutate some DNA!\n", purple, reset)
		case "I|AM|PHYSICS":
			simulationTools = core.CreateTools(connection, connectionReader)
			fmt.Printf("🚀 %s[PHYSICS ALIGNED]%s Trajectory processing unit online. Gravity enabled!\n", cyan, reset)
		case "I|AM|VISUALIZATION":
			visualizationTools = core.CreateTools(connection, connectionReader)
			fmt.Printf("🎨 %s[PYTHON APU ONLINE]%s Visual rendering engine attached. Let's make it look clean!\n", colorGreen, reset)
		case "I|AM|PRINTER":
			printerTools = core.CreateTools(connection, connectionReader)
			fmt.Printf("🖨️ %s[PRINTER READY]%s Logging node linked up. Prepare your papers, we are shifting into print mode!\n", bold, reset)
		}
	}

	fmt.Println("\n🤩 " + colorGreen + "════════ NET ORCHESTRA ASSEMBLED! Let's start the show! ════════" + reset)

	defer func() {
		if inputTools.Conn != nil {
			_ = inputTools.Writer.Flush()
			inputTools.Conn.Close()
		}
		if geneticTools.Conn != nil {
			_ = geneticTools.Writer.Flush()
			geneticTools.Conn.Close()
		}
		if simulationTools.Conn != nil {
			_ = simulationTools.Writer.Flush()
			simulationTools.Conn.Close()
		}
		if visualizationTools.Conn != nil {
			_ = visualizationTools.Writer.Flush()
			visualizationTools.Conn.Close()
		}
		if printerTools.Conn != nil {
			_ = printerTools.Writer.Flush()
			printerTools.Conn.Close()
		}

		fmt.Println("\n🛌 ────────────── SERVER IS GOING TO SLEEP ──────────────")
		fmt.Printf("😴 %s[DISCONNECTING TUNNELS]%s All cluster nodes freed. Session cleared safely.\n", colorGreen, reset)
		fmt.Println("💤 Simulation destroyed. Going offline... See you next launch! zZz...")
		fmt.Println("───────────────────────────────────────────────────")
	}()

	defer finishWorking(geneticTools.Writer, simulationTools.Writer, visualizationTools.Writer, printerTools.Writer)

	targetX, targetY, targetZ, populationSize, generations, firstRocket, maximumBurnRate, err := readAllInput(inputTools.Reader)
	if err != nil {
		fmt.Printf("😤 %s[DATA CORRUPTION]%s The parser sent completely broken metrics instead of numbers: %v\n", red, reset, err)
	}

	_, err = fmt.Fprintf(printerTools.Writer, "CHARACTERISTICS|%v|%v|%v|%v\n", generations, targetX, targetY, targetZ)
	if err != nil {
		fmt.Printf("😠 %s[PRINTER FAULT]%s Failed to pipe initial simulation specs to the printer unit: %v\n", red, reset, err)
		return
	}
	printerTools.Writer.Flush()

	_, err = fmt.Fprintf(visualizationTools.Writer, "TARGET|%v|%v|%v\n", targetX, targetY, targetZ)
	if err != nil {
		fmt.Printf("😡 %s[NETWORK DISASTER]%s Couldn't stream target locked coordinates to Python engine: %v\n", red, reset, err)
		return
	}

	rocketPopulation, err := startFirstAlgorithm(geneticTools.Reader, geneticTools.Writer, firstRocket, populationSize, maximumBurnRate)
	if err != nil {
		fmt.Printf("😰 %s[ALGORITHM FAILURE]%s Core engine broke on initial DNA sequence generation: %v\n", red, reset, err)
	}

	medianHistory := make([]float64, medianHisorySize)
	adaptiveThreshold := 0.0

	for countGenerations := 0; countGenerations < generations; countGenerations++ {
		fmt.Printf("\n⏳ %s[PROCESSING]%s Running natural selection loops: Gen %s%d/%d%s...\n",
			cyan, reset, purple, countGenerations+1, generations, reset)

		medianNextFitness, bestRockets2, err := startCalculation(simulationTools.Reader, simulationTools.Writer, targetX, targetY, targetZ, populationSize, rocketPopulation)
		if countGenerations == 0 {
			initialMedian := medianNextFitness
			adaptiveThreshold = math.Max(80.0, initialMedian*0.005)
		}

		medianHistory[(countGenerations)%medianHisorySize] = medianNextFitness

		_, err = fmt.Fprintf(printerTools.Writer, "MEDIAN|%v\n", medianNextFitness)
		if err != nil {
			fmt.Printf("🤷 %s[METRIC LOSS]%s Eh... the median values just slipped out of my hands on the way to the printer: %v\n", yellow, reset, err)
			return
		}
		printerTools.Writer.Flush()

		var bestRocket core.Rocket
		for id, rocket := range bestRockets2 {
			if id == 0 || rocket.Fitness < bestRocket.Fitness {
				bestRocket = rocket
			}
			if rocket.TrajectoryCSV == "" {
				rocket.TrajectoryCSV = "0.00,0.00,0.00"
			}

			_, err = fmt.Fprintf(visualizationTools.Writer, "ROCKET_ID|%v|%v\n", id+1, rocket.TrajectoryCSV)
			if err != nil {
				fmt.Printf("🤬 %s[TRANSMISSION CRASH]%s Severe payload drop while routing flight charts: %v\n", red, bold, err)
				return
			}
			visualizationTools.Writer.Flush()

			if rocket.Fitness <= adaptiveThreshold && rocket.Y > 5 {
				_, err = fmt.Fprintf(printerTools.Writer, "CHAMPION|%v|%v|%v|%v|%v|%v|%v|%v|%v\n", rocket.PitchDegree, rocket.YawDegree, rocket.Fuel, rocket.BurnRate, rocket.BodyMass, rocket.X, rocket.Y, rocket.Z, rocket.Fitness)
				if err != nil {
					fmt.Printf("🤦 %s[VICTORY BLUNDER]%s You won't believe it! We literally hit the target, but the printer choked on the final applause: %v\n", red, bold, err)
					return
				}
				printerTools.Writer.Flush()
				fmt.Println("\n🎉🎉🎉 " + colorGreen + "══════════════════ OOOOH YEAH! WE HIT THE TARGET! ══════════════════" + reset + " 🎉🎉🎉")
				fmt.Printf("🏆 %s[EUREKA!]%s Perfect trajectory optimized at Generation %s#%d%s!\n", colorGreen, reset, purple, countGenerations+1, reset)
				fmt.Printf("📊 Champion Specs:\n")
				fmt.Printf("   • Launch Angle :  %s%.2f°%s\n", cyan, rocket.PitchDegree, reset)
				fmt.Printf("   • Fuel Load    :  %s%.2f kg%s\n", cyan, rocket.Fuel, reset)
				fmt.Printf("   • Fuel Burn    :  %s%.2f kg/s%s\n", cyan, rocket.BurnRate, reset)
				fmt.Printf("--------------------------------------------------\n")
				fmt.Printf("📈 Flight Summary Matrix:\n")
				fmt.Printf("   • Final X Pos  :  %s%.2f meters%s\n", colorGreen, rocket.X, reset)
				fmt.Printf("   • Max Altitude :  %s%.2f meters%s\n", purple, rocket.Y, reset)
				fmt.Printf("   • Target Miss  :  %s%.2f meters%s (Threshold Limit: %.2f)\n", yellow, rocket.Fitness, reset, adaptiveThreshold)
				fmt.Println(colorGreen + "══════════════════════════════════════════════════════════" + reset)

				fmt.Println("\n👋 [SYSTEM] Mission completed successfully. Initializing orderly power-down sequence...")
				time.Sleep(1 * time.Second)
				return
			}
		}

		_, err = fmt.Fprintln(visualizationTools.Writer, "END_OF_GENERATION")
		if err != nil {
			fmt.Printf("an error occured: %v", err)
		}
		visualizationTools.Writer.Flush()

		_, err = fmt.Fprintf(printerTools.Writer, "GENERATION_CHAMPION|%v|%v|%v|%v|%v|%v|%v|%v|%v\n", bestRocket.PitchDegree, bestRocket.YawDegree, bestRocket.Fuel, bestRocket.BurnRate, bestRocket.BodyMass, bestRocket.X, bestRocket.Y, bestRocket.Z, bestRocket.Fitness)
		if err != nil {
			fmt.Printf("🙄 %s[PRINTER SLACKING]%s The generation leader specs didn't go through. Whatever, Python will survive without it: %v\n", yellow, reset, err)
			return
		}
		printerTools.Writer.Flush()

		err = waitSlowPython(visualizationTools.Reader)
		if err != nil {
			fmt.Printf("❌ %s[CRITICAL PIPELINE SYNC]%s Python framework locked up on Gen %d: %v\n", red, reset, countGenerations+1, err)
			return
		}
		rocketPopulation, err = startAlgorithm(geneticTools.Reader, geneticTools.Writer, bestRockets2, populationSize, countGenerations+1, medianHistory)
		if err != nil {
			fmt.Printf("😰 %s[ALGORITHM FAILURE]%s DNA crossover error occurred on Gen %d: %v\n", red, reset, countGenerations+1, err)
		}
	}

	fmt.Printf("\n😮 %s[BOUNDARIES MET]%s Generation limits exhausted, target threshold not hit. Well, at least we tried our best!\n", yellow, reset)
}
