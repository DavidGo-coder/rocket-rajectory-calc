# 🚀 Rocket Trajectory GA Simulator (v1.0.0)

Highly optimized, real-time ballistic physics simulation engine paired with an **Evolutionary Artificial Intelligence (Genetic Algorithm)**. Built with **Go** for high-performance concurrent mathematical computations and **Python** for real-time telemetry stream rendering.

---

## 🛠️ Tech Stack & Architecture

### 🚄 Go Computational Core (Backend)
- **High-Performance Physics:** Real-time integration of differential flight equations via the **Euler-Cromer method** (dt = 0.01s).
- **Network I/O Pipeline:** Low-level, non-blocking asynchronous **TCP Socket Server** (:8080) for streaming live telemetry frames.
- **Concurrency Model:** Multi-threaded parallel processing using Go-routines (`go func`) and thread-safe channels (`chan`) to evaluate entire populations instantly.

### 📡 Python Radar Interface (Frontend)
- **Dynamic Telemetry Stream:** Real-time raw socket data ingestion and data packet decoding.
- **3D Graphics Engine:** Advanced trajectory reconstruction, spatial tracking, and evolutionary convergence analysis using **Matplotlib / Pandas**.

---

## 📈 Project Status & Release Roadmap

- [x] **v1.0.0 — Initial Stable Release**
  - [x] High-fidelity multi-threaded physical simulation engine in Go (Gravity gradients & Aerodynamic drag).
  - [x] Integration of the Evolutionary Genetic Algorithm (GA) with 5-gene chromosome DNA.
  - [x] Anti-cheat safety guardrails preventing negative mass or unphysical fuel ratios (\(m_{fuel} \le 10 \cdot m_{dry}\)).
  - [x] Real-time TCP network bridge linking Go & Python pipelines.
  - [x] Macro-mutation engine (`ExclusiveChild`) with dynamic scale adaptation to break local minima stagnation.
- [ ] **v2.0.0 — Future Enhancements**
  - [ ] Implementation of full 3D spatial flight dynamics using the `YawDegree` gene.
  - [ ] Dynamic wind vectors and true coriolis effect calculations.

---

## 🔬 Physics Simulation Engine Specifications

The backend solver evaluates individual rocket trajectories by accounting for continuous real-world forces:
1. **Variable System Mass:** Modeled iteratively as \(m(t) = m_{dry} + m_{fuel}(t)\) influenced by the fuel consumption rate (\(\dot{m}\), `BurnRate`).
2. **Atmospheric Drag Profile:** Dynamic barometric drag coefficient scaling exponentially with altitude:
   \[F_{drag} = C_x \cdot V^2 \cdot e^{-\frac{y}{8500}}\]
3. **Gravitational Gradient:** Real-time adjustments to gravitational acceleration (\(g\)) based on distance from the planet's core (Newton's law of universal gravitation):
   \[g(y) = g_0 \cdot \left(\frac{R_E}{R_E + y}\right)^2\]

---

## 🧬 Evolutionary Optimization (Genetic Algorithm)

Instead of a brute-force approach, the core engine leverages an advanced **Genetic Algorithm (GA)** to find the optimal structural design and launch parameters.

### Chromosome Map (Rocket DNA):
- **Pitch Angle** (`PitchDegree`, °) — Launch elevation vector.
- **Yaw Angle** (`YawDegree`, °) — Azimuth vector.
- **Initial Fuel Mass** (`Fuel`, kg) — Total chemical energy storage.
- **Engine Burn Rate** (`BurnRate`, kg/s) — Thrust-to-weight controller.
- **Engine Efficiency** (`EngineEfficiency`, m/s) — Automated pre-processed specific impulse (\(I_{sp}\)).

### Fitness Evaluation Function:
The multi-criteria fitness engine evaluates performance through:
- **Spatial Precision:** Minimizing the absolute Euclidean distance between the trajectory arc and the target coordinate vector.
- **Fuel Economy Bonus:** If a specimen successfully enters the target zone, it receives a fitness multiplier for conserved fuel mass, training the AI to select the most resource-efficient engineering solutions.

---

## 🚀 How to Run the Ecosystem

1. Initialize the Python telemetry listener pipeline to prepare the 3D visual canvas:
   ```bash
   python radar.py
   ```
2. Build and launch the concurrent Go evolutionary engine:
   ```bash
   go run main.go
   ```
3. Input target spatial coordinates (e.g., `10000, 0, 5`) and design the core DNA parameters of your baseline specimen to watch the algorithm converge across generations.