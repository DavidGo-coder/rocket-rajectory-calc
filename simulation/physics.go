package main

import (
	"bufio"
	"fmt"
	"math"
	"net"
	"sort"
	"strings"
	"sync"

	core "github.com/DavidGo-coder/rocket-trajecory-calc"
)

const (
	portConnection string = "8800"
	hostConnection string = "127.0.0.1"
	colorReset            = "\x1b[0m"
	colorRed              = "\x1b[31m"
	colorGreen            = "\x1b[32m"
	colorYellow           = "\x1b[33m"
	colorBlue             = "\x1b[34m"
	colorCyan             = "\x1b[36m"
	colorBold             = "\x1b[1m"
)

func readCalculationParameters(listCharacerisics []string) (int, int, int, int, int, error) {
	targetX, err := core.ParseInt(listCharacerisics, 0)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	targetY, err := core.ParseInt(listCharacerisics, 1)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	targetZ, err := core.ParseInt(listCharacerisics, 2)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	populationSize, err := core.ParseInt(listCharacerisics, 3)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	length, err := core.ParseInt(listCharacerisics, 4)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}

	return targetX, targetY, targetZ, populationSize, length, nil
}

func CalculationTrajectory(targetCoordinateX, targetCoordinateY, targetCoordinateZ, populationSize int, rocketPopulation []core.Rocket) ([]core.Rocket, float64) {
	if len(rocketPopulation) < populationSize {
		populationSize = len(rocketPopulation)
	}
	if populationSize == 0 {
		fmt.Println("⚠️  \x1b[33m[PHYSICS WARNING]\x1b[0m Empty population received for simulation. Skipping.")
		return nil, 0.0
	}

	const (
		colorReset  = "\x1b[0m"
		colorCyan   = "\x1b[36m"
		colorGreen  = "\x1b[32m"
		colorYellow = "\x1b[33m"
		colorPurple = "\x1b[35m"
	)

	fmt.Printf("🚀 %s[PHYSICS]%s Starting simulation for %s%d%s rockets targeting (%d, %d)...\n",
		colorCyan, colorReset, colorPurple, populationSize, colorReset, targetCoordinateX, targetCoordinateY)

	var wg sync.WaitGroup
	wg.Add(populationSize)

	type ResultSimulation struct {
		x             float64
		y             float64
		trajectoryCSV string
		score         float64
		resultIndex   int
	}

	channelResultSimulation := make(chan ResultSimulation, populationSize)

	for countSimulation := 0; countSimulation < populationSize; countSimulation++ {
		go func(index int, fuel, pitchDegree, yawDegree, burnRate, engineEfficiency, bodyMass float64) {
			defer wg.Done()

			pitchRadians := pitchDegree * math.Pi / 180.0
			vx, vy := 0.0, 0.0
			x, y := 0.0, 0.0
			time := 0.0
			maxY := 0.0

			var trajectoryPoints []string
			stepCounter := 0

			targetX := float64(targetCoordinateX)
			targetY := float64(targetCoordinateY)

			minDistToTarget := math.Sqrt(targetX*targetX + targetY*targetY)

			for time < 275.0 {
				currentMass := bodyMass + fuel
				if currentMass < 0.1 {
					currentMass = 0.1
				}

				var thrustForceX, thrustForceY float64
				if fuel > 0 {
					dynamicThrust := burnRate * engineEfficiency

					fuelConsumed := burnRate * core.OneMomentSimulation
					if fuelConsumed > fuel {
						dynamicThrust = (fuel / core.OneMomentSimulation) * engineEfficiency
						fuel = 0
					} else {
						fuel -= fuelConsumed
					}

					thrustForceX = dynamicThrust * math.Cos(pitchRadians)
					thrustForceY = dynamicThrust * math.Sin(pitchRadians)
				}

				var fdragX, fdragY float64
				v := math.Sqrt(vx*vx + vy*vy)
				if v > 0.001 {
					currentDragCoef := core.AirResistance * math.Exp(-y/8500.0)
					fdragTotal := currentDragCoef * v * v

					fdragX = fdragTotal * (vx / v)
					fdragY = fdragTotal * (vy / v)
				}

				t := core.RadiusPlanetEarth / (core.RadiusPlanetEarth + y)
				currentGravitation := core.GravitationalConstant * (t * t)

				ax := (thrustForceX - fdragX) / currentMass
				ay := ((thrustForceY - fdragY) / currentMass) - currentGravitation

				x += vx * core.OneMomentSimulation
				y += vy * core.OneMomentSimulation
				vx += ax * core.OneMomentSimulation
				vy += ay * core.OneMomentSimulation

				if y <= 0 && vy < 0 {
					y = 0
					if time < 0.5 || thrustForceY > currentGravitation {
						vy = 0
					} else {
						x -= vx * core.OneMomentSimulation
						vx = 0
						vy = 0
						trajectoryPoints = append(trajectoryPoints, fmt.Sprintf("%.2f,%.2f", x, y))
						break
					}
				}

				if y > maxY {
					maxY = y
				}

				dx := targetX - x
				dy := targetY - y
				currentDist := math.Sqrt(dx*dx + dy*dy)

				if currentDist < minDistToTarget {
					minDistToTarget = currentDist
				}

				if stepCounter%50 == 0 {
					trajectoryPoints = append(trajectoryPoints, fmt.Sprintf("%.2f,%.2f", x, y))
				}
				stepCounter++
				time += core.OneMomentSimulation
			}

			efficiencyScore := minDistToTarget
			if maxY < 5.0 {
				efficiencyScore += core.FailedLaunchPenalty
			} else {
				proximity := math.Exp(-minDistToTarget / 500.0)
				efficiencyScore -= fuel * 2.0 * proximity
			}

			channelResultSimulation <- ResultSimulation{
				x:             x,
				y:             maxY,
				score:         efficiencyScore,
				trajectoryCSV: strings.Join(trajectoryPoints, ";"),
				resultIndex:   index,
			}
		}(countSimulation, rocketPopulation[countSimulation].Fuel, rocketPopulation[countSimulation].PitchDegree, rocketPopulation[countSimulation].YawDegree, rocketPopulation[countSimulation].BurnRate, rocketPopulation[countSimulation].EngineEfficiency, rocketPopulation[countSimulation].BodyMass)
	}
	wg.Wait()
	close(channelResultSimulation)

	for result := range channelResultSimulation {
		rocketPopulation[result.resultIndex].Fitness = result.score
		rocketPopulation[result.resultIndex].X = result.x
		rocketPopulation[result.resultIndex].Y = result.y
		rocketPopulation[result.resultIndex].TrajectoryCSV = result.trajectoryCSV
	}

	sort.Slice(rocketPopulation[:populationSize], func(i, j int) bool {
		return rocketPopulation[i].Fitness < rocketPopulation[j].Fitness
	})

	// --- КРАСИВЫЙ И ИНФОРМАТИВНЫЙ ВЫВОД ---
	medianFitness := rocketPopulation[populationSize/2].Fitness
	bestFitness := rocketPopulation[0].Fitness
	bestMaxY := rocketPopulation[0].Y

	fmt.Println("─────────────── Generation Summary ───────────────")
	fmt.Printf("🧬 %s[EVOLUTION]%s Median Miss: %s%.2f meters%s\n",
		colorCyan, colorReset, colorYellow, medianFitness, colorReset)
	fmt.Printf("🏆 %s[TOP ROCKET]%s Best Score: %s%.2f%s | Max Altitude: %s%.1f m%s\n",
		colorGreen, colorReset, colorGreen, bestFitness, colorReset, colorPurple, bestMaxY, colorReset)
	fmt.Println("──────────────────────────────────────────────────")

	var bestRockets []core.Rocket
	for b := 0; b < 10 && b < populationSize; b++ {
		bestRockets = append(bestRockets, rocketPopulation[b])
	}

	return bestRockets, medianFitness
}

func main() {
	address := net.JoinHostPort(hostConnection, portConnection)

	fmt.Printf("🌐 %s[NETWORK]%s Attempting to connect to remote host at %s...\n", colorCyan, colorReset, address)
	connection, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Printf("❌ %s%s[CONNECTION ERROR]%s Could not establish connection to remote host at %v: %v\n",
			colorRed, colorBold, colorReset, address, err)
		return
	}
	defer connection.Close()
	fmt.Printf("✅ %s[CONNECTED]%s Successfully established TCP tunnel with server.\n", colorGreen, colorReset)

	reader := bufio.NewReader(connection)
	tools := core.CreateTools(connection, reader)

	defer func() {
		if err := tools.Writer.Flush(); err != nil {
			fmt.Printf("❌ %s[IO ERROR]%s Failed to flush final data buffer to network: %v\n", colorRed, colorReset, err)
		} else {
			fmt.Printf("✨ %s[SUCCESS]%s Data synchronization completed successfully. Connection closed cleanly.\n", colorGreen, colorReset)
		}
	}()

	_, err = fmt.Fprintln(tools.Writer, "I|AM|PHYSICS")
	if err != nil {
		fmt.Printf("❌ %s[HANDSHAKE FAILED]%s Protocol verification message rejected by server: %v\n", colorRed, colorReset, err)
		return
	}
	tools.Writer.Flush()
	fmt.Printf("🤝 %s[HANDSHAKE]%s Sent identity 'I|AM|PHYSICS' to remote server.\n", colorBlue, colorReset)

	for {
		obtainedCommand, err := core.ReadInput(tools.Reader)
		if err != nil {
			fmt.Printf("❌ %s[NETWORK DISCONNECT]%s Connection lost or closed by remote host: %v\n", colorRed, colorReset, err)
			return
		}

		switch obtainedCommand {
		case "COMMAND1":
			listCharacerisics, err := core.ReadInputSliced(tools.Reader)
			if err != nil {
				fmt.Printf("❌ %s[DATA ERROR]%s Failed to read simulation configuration parameters: %v\n", colorRed, colorReset, err)
				continue
			}

			targetX, targetY, targetZ, populationSize, length, err := readCalculationParameters(listCharacerisics)
			if err != nil {
				fmt.Printf("❌ %s[PARSING ERROR]%s Invalid format in simulation parameters: %v\n", colorRed, colorReset, err)
				continue
			}

			fmt.Printf("📥 %s[COMMAND1]%s Received config. Target: (%d, %d, %d) | Pop Size: %d\n",
				colorBlue, colorReset, targetX, targetY, targetZ, populationSize)

			rocketPopulation, err := core.ParseNRocket(length, tools.Reader)
			if err != nil {
				fmt.Printf("❌ %s[PARSE ERROR]%s Failed to parse population DNA: %v\n", colorRed, colorReset, err)
				continue
			}

			bestRockets, median := CalculationTrajectory(targetX, targetY, targetZ, populationSize, rocketPopulation)

			_, err = fmt.Fprintf(tools.Writer, "%v|%v\n", median, len(bestRockets))
			if err != nil {
				fmt.Printf("❌ %s[TRANSMISSION ERROR]%s Failed to send generational median score to server: %v\n", colorRed, colorReset, err)
			}
			tools.Writer.Flush()

			for i := 0; i < len(bestRockets); i++ {
				currentRocket := bestRockets[i]
				_, err := fmt.Fprintf(tools.Writer, "%v|%v|%v|%v|%v|%v|%v|%v|%v|%v|%v\n", currentRocket.PitchDegree, currentRocket.YawDegree, currentRocket.Fuel, currentRocket.BurnRate, currentRocket.BodyMass, currentRocket.EngineEfficiency, currentRocket.X, currentRocket.Y, currentRocket.Z, currentRocket.Fitness, currentRocket.TrajectoryCSV)
				if err != nil {
					fmt.Printf("❌ %s[TRANSMISSION ERROR]%s Failed to stream optimized rocket DNA payload back to server: %v\n", colorRed, colorReset, err)
					break
				}
				tools.Writer.Flush()
			}
			fmt.Printf("📤 %s[TRANSMISSION]%s Successfully sent back median (%.2f) and %d best rocket DNA sets.\n",
				colorGreen, colorReset, median, len(bestRockets))

		case "FINISHWORKING":
			fmt.Printf("🛑 %s[SHUTDOWN]%s Received termination command from server. Exiting...\n", colorYellow, colorReset)
			return
		default:
			fmt.Printf("⚠️  %s[PROTOCOL WARNING]%s Received unknown command sequence from server: %s\n", colorYellow, colorReset, obtainedCommand)
			continue
		}
	}
}
