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
	totalProgram             = 4

	colorReset  = "\x1b[0m"
	colorRed    = "\x1b[31m"
	colorGreen  = "\x1b[32m"
	colorYellow = "\x1b[33m"
	colorBlue   = "\x1b[34m"
	colorPurple = "\x1b[35m"
	colorCyan   = "\x1b[36m"
	colorBold   = "\x1b[1m"
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

func finishWorking(writer1 *bufio.Writer, writer2 *bufio.Writer, writer3 *bufio.Writer) error {
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
	return nil
}

func waitSlowPython(bufferedReader *bufio.Reader) error {
	currentTime := time.Now()

	_, err := bufferedReader.ReadString('\n')
	if err != nil && err != io.EOF {
		fmt.Printf("🤬 %s[PYTHON CRASHED]%s Oh come on! The visualization client just walked out on us: %v\n", colorRed, colorBold, err)
		return err
	}

	if time.Since(currentTime) < timeOneGeneration {
		timeSpent := time.Since(currentTime)
		fmt.Printf("🥱 %s[WAITING FOR PYTHON]%s Calculations are too fast, Python is lagging behind... Napping for %v\n", colorYellow, colorReset, timeOneGeneration-timeSpent)
		time.Sleep(timeOneGeneration - timeSpent)
	}
	fmt.Printf("✨ %s[SYNC SUCCESS]%s Python finally caught up! Generation data rendered beautifully! 🎉\n", colorGreen, colorReset)
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

	for i := 0; i < len(medianHistory); i++ {
		if i != 0 {
			_, err = fmt.Fprint(bufferedWriter, "|")
			if err != nil {
				return []core.Rocket{}, err
			}
			bufferedWriter.Flush()
		}
		_, err = fmt.Fprintf(bufferedWriter, "%v", medianHistory[i])
		if err != nil {
			return []core.Rocket{}, err
		}
		bufferedWriter.Flush()
	}

	_, err = fmt.Fprintf(bufferedWriter, "\n%v|%v\n", countGenerations, len(bestRockets))
	if err != nil {
		return []core.Rocket{}, err
	}
	bufferedWriter.Flush()

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

	listener, err := net.Listen("tcp", ":"+portConnection)
	if err != nil {
		fmt.Printf("🤬 %s[LAUNCH CRASH]%s Can't bind port %v! Did someone steal it?! Error: %v\n", colorRed, colorBold, portConnection, err)
		return
	}

	fmt.Printf("📡 %s[I AM ALIVE!]%s TCP Server successfully bound to port %s%v%s. Awaiting guests...\n", colorCyan, colorReset, colorPurple, portConnection, colorReset)
	defer listener.Close()

	for countProgram := 0; countProgram < totalProgram; countProgram++ {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Printf("😭 %s[HANDSHAKE ABORTED]%s A module tried to link up but crashed down: %v\n", colorRed, colorReset, err)
		}
		connectionReader := bufio.NewReader(connection)

		obtainedData, err := connectionReader.ReadString('\n')
		if err != nil {
			fmt.Printf("🧐 %s[WHAT'S THAT NOISE?]%s Received totally unreadable garbage bytes: %v\n", colorYellow, colorReset, err)
			return
		}
		obtainedData = strings.TrimSpace(obtainedData)

		switch obtainedData {
		case "I|AM|PARSER":
			inputTools = core.CreateTools(connection, connectionReader)
			fmt.Printf("📥 %s[HELLO, PARSER!]%s Input module has entered the lobby. Config specs received!\n", colorBlue, colorReset)
		case "I|AM|ALGORITHM":
			geneticTools = core.CreateTools(connection, connectionReader)
			fmt.Printf("🧬 %s[EVOLUTION IS HERE]%s Genetic algorithm hooked up. Ready to mutate some DNA!\n", colorPurple, colorReset)
		case "I|AM|PHYSICS":
			simulationTools = core.CreateTools(connection, connectionReader)
			fmt.Printf("🚀 %s[PHYSICS ALIGNED]%s Trajectory processing unit online. Gravity enabled!\n", colorCyan, colorReset)
		case "I|AM|VISUALIZATION":
			visualizationTools = core.CreateTools(connection, connectionReader)
			fmt.Printf("🎨 %s[PYTHON APU ONLINE]%s Visual rendering engine attached. Let's make it look clean!\n", colorGreen, colorReset)
		}
	}

	fmt.Println("\n🤩 " + colorGreen + "════════ NET ORCHESTRA ASSEMBLED! Let's start the show! ════════" + colorReset)

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

		fmt.Println("\n🛌 ────────────── SERVER IS GOING TO SLEEP ──────────────")
		fmt.Printf("😴 %s[DISCONNECTING TUNNELS]%s All cluster nodes freed. Session cleared safely.\n", colorGreen, colorReset)
		fmt.Println("💤 Simulation destroyed. Going offline... See you next launch! zZz...")
		fmt.Println("───────────────────────────────────────────────────")
	}()

	defer finishWorking(geneticTools.Writer, simulationTools.Writer, visualizationTools.Writer)

	targetX, targetY, targetZ, populationSize, generations, firstRocket, maximumBurnRate, err := readAllInput(inputTools.Reader)
	if err != nil {
		fmt.Printf("😤 %s[DATA CORRUPTION]%s The parser sent completely broken metrics instead of numbers: %v\n", colorRed, colorReset, err)
	}

	_, err = fmt.Fprintf(visualizationTools.Writer, "TARGET|%v,%v\n", targetX, targetY)
	if err != nil {
		fmt.Printf("😡 %s[NETWORK DISASTER]%s Couldn't stream target locked coordinates to Python engine: %v\n", colorRed, colorReset, err)
		return
	}
	visualizationTools.Writer.Flush()

	rocketPopulation, err := startFirstAlgorithm(geneticTools.Reader, geneticTools.Writer, firstRocket, populationSize, maximumBurnRate)
	if err != nil {
		fmt.Printf("😰 %s[ALGORITHM FAILURE]%s Core engine broke on initial DNA sequence generation: %v\n", colorRed, colorReset, err)
	}

	medianHistory := make([]float64, medianHisorySize)
	adaptiveThreshold := 0.0

	for countGenerations := 0; countGenerations < generations; countGenerations++ {
		fmt.Printf("\n⏳ %s[PROCESSING]%s Running natural selection loops: Gen %s%d/%d%s...\n",
			colorCyan, colorReset, colorPurple, countGenerations+1, generations, colorReset)

		medianNextFitness, bestRockets2, err := startCalculation(simulationTools.Reader, simulationTools.Writer, targetX, targetY, targetZ, populationSize, rocketPopulation)
		if countGenerations == 0 {
			initialMedian := medianNextFitness
			adaptiveThreshold = math.Max(80.0, initialMedian*0.005)
		}

		medianHistory[(countGenerations)%medianHisorySize] = medianNextFitness

		for id, rocket := range bestRockets2 {
			if rocket.TrajectoryCSV == "" {
				rocket.TrajectoryCSV = "0.00,0.00"
			}

			_, err = fmt.Fprintf(visualizationTools.Writer, "ROCKET_ID:%v;%v\n", id+1, rocket.TrajectoryCSV)
			if err != nil {
				fmt.Printf("🤬 %s[TRANSMISSION CRASH]%s Severe payload drop while routing flight charts: %v\n", colorRed, colorBold, err)
				return
			}
			visualizationTools.Writer.Flush()

			if rocket.Fitness <= adaptiveThreshold && rocket.Y > 5 {
				fmt.Println("\n🎉🎉🎉 " + colorGreen + "══════════════════ OOOOH YEAH! WE HIT THE TARGET! ══════════════════" + colorReset + " 🎉🎉🎉")
				fmt.Printf("🏆 %s[EUREKA!]%s Perfect trajectory optimized at Generation %s#%d%s!\n", colorGreen, colorReset, colorPurple, countGenerations+1, colorReset)
				fmt.Printf("📊 Champion Specs:\n")
				fmt.Printf("   • Launch Angle :  %s%.2f°%s\n", colorCyan, rocket.PitchDegree, colorReset)
				fmt.Printf("   • Fuel Load    :  %s%.2f kg%s\n", colorCyan, rocket.Fuel, colorReset)
				fmt.Printf("   • Fuel Burn    :  %s%.2f kg/s%s\n", colorCyan, rocket.BurnRate, colorReset)
				fmt.Printf("--------------------------------------------------\n")
				fmt.Printf("📈 Flight Summary Matrix:\n")
				fmt.Printf("   • Final X Pos  :  %s%.2f meters%s\n", colorGreen, rocket.X, colorReset)
				fmt.Printf("   • Max Altitude :  %s%.2f meters%s\n", colorPurple, rocket.Y, colorReset)
				fmt.Printf("   • Target Miss  :  %s%.2f meters%s (Threshold Limit: %.2f)\n", colorYellow, rocket.Fitness, colorReset, adaptiveThreshold)
				fmt.Println(colorGreen + "══════════════════════════════════════════════════════════" + colorReset)

				fmt.Println("\n👋 [SYSTEM] Mission completed successfully. Initializing orderly power-down sequence...")
				time.Sleep(1 * time.Second)
				return
			}
		}

		err = waitSlowPython(visualizationTools.Reader)
		if err != nil {
			fmt.Printf("❌ %s[CRITICAL PIPELINE SYNC]%s Python framework locked up on Gen %d: %v\n", colorRed, colorReset, countGenerations+1, err)
			return
		}
		rocketPopulation, err = startAlgorithm(geneticTools.Reader, geneticTools.Writer, bestRockets2, populationSize, countGenerations+1, medianHistory)
		if err != nil {
			fmt.Printf("😰 %s[ALGORITHM FAILURE]%s DNA crossover error occurred on Gen %d: %v\n", colorRed, colorReset, countGenerations+1, err)
		}
	}

	fmt.Printf("\n😮 %s[BOUNDARIES MET]%s Generation limits exhausted, target threshold not hit. Well, at least we tried our best!\n", colorYellow, colorReset)
}
