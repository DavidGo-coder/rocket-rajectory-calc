# 🚀 Rocket Trajectory Genetic Optimization Engine

A high-performance, concurrent **Go-based evolutionary simulation engine** designed to optimize rocket launch DNA parameters. The ecosystem couples an **Euler-Cromer physics engine** with a **Genetic Algorithm (GA)**, streaming real-time flight trajectories over a synchronized **TCP bridge** to a Python-based visual telemetry canvas (`radar.py`).

---

## 📈 Project Status & Release Roadmap

- [x] **v1.0.0 — Monolithic Prototype**
  - [x] High-fidelity physical simulation engine in Go (gravity gradients & aerodynamic drag).
  - [x] Evolutionary Genetic Algorithm (GA) with 5-gene chromosome DNA.
  - [x] Real-time TCP network bridge linking Go & Python pipelines.
- [x] **v1.1.0 — Architecture Optimization & Bugfixes**
  - [x] **Modular Monolith:** Separated core concerns by splitting the single file into isolated packages (`simulation`, `genetic`).
  - [x] **Population Guardrails:** Patched critical out-of-range panics and corrected initial mutation cycle overrides.
- [x] **v1.2.0 — Pipeline Synchronization & Core Stability (Current)**
  - [x] **Deterministic Frame Sync:** Enhanced the runtime orchestrator (`main.go`) with a strict 3000ms minimum window constraint per generation (`WaitSlowPython`), preventing Python telemetry UI frame drops.
  - [x] **Adaptive Convergence Thresholds:** Integrated dynamic stop-conditions based on historical generation performance.
  - [x] **Advanced Code Decoupling:** Re-architected entry points to cleanly interface with mathematical validation layers.
- [ ] **v2.0.0 — Full 3D Spatial Vector Dynamics (Target: ~2027)**
  - [ ] Refactoring the core simulation from 2D vector space to full 3D coordinate system processing.
  - [ ] Activating the compiled `YawDegree` gene inside active multi-threaded physics matrices.
  - [ ] Introducing dynamic crosswind vectors and true Coriolis effect calculations.

---

## 🏗️ Architectural Layout

The repository enforces the **Single Responsibility Principle** to guarantee clean isolation between data models, visualization layers, and optimization routines:

```text
rocketBehavior/
├── go.mod                # Core Go Module Definition
├── main.go               # Orchestrator, CLI UI Controller & TCP Socket Server
├── simulation/
│   ├── models.go         # Shared Aerospace Entities (Rocket Struct & Constants)
│   └── physics.go        # Concurrent Euler-Cromer Trajectory Engine (sync.WaitGroup)
└── genetic/
    └── algorithm.go      # DNA Validation, Crossover, and Mutation Mechanics
```

---

## 🏎️ Core Technical Mechanics

### 1. Aerospace Physics Simulation
Flight profiles are integrated using **Euler-Cromer math** with a step resolution of Δ t = 0.01 seconds (`OneMomentSimulation`). The calculation pipeline accurately tracks:
* **Mass Depletion Burn:** Fuel decreases dynamically by `BurnRate * dt`. Total thrust force combines mass flow rate with engine efficiency calculations.
* **Atmospheric Drag Gradient:** Air resistance scales non-linearly with velocity and decays exponentially relative to altitude: 
  \[F_{\text{drag}} = \text{AirResistance} \cdot e^{-\frac{y}{8500}} \cdot v^2\]
* **Gravity Gradient:** Earth's gravitational acceleration decays with altitude following Newton's inverse-square law:
  \[g(y) = g_0 \cdot \left(\frac{R_E}{R_E + y}\right)^2\]

> 🌐 **Note on 3D Transition State:** In v1.2.0, the UI proactively captures 3D targets (`X, Y, Z`) and genetic sequences mutate the `YawDegree` gene to establish the foundation for our upcoming spatial engine. Currently, core trajectory calculations project onto the 2D plane as the framework transitions toward full 3D physics integration.

### 2. Genetic Optimization & Fitness Scaling
The optimization engine minimizes a complex **Fitness Score** (target miss distance) using continuous multi-gene chromosomes:
* **Launch Failure Penalties:** If a rocket fails to clear an apogee (\(Y_{\text{max}}\)) of at least **5.0 meters**, it incurs a flat **+50,000 penalty score** (`FailedLaunchPenalty`).
* **Fuel Efficiency Incentives:** Successful flights receive a fuel retention bonus scaled by target proximity: \(\text{Score} = \text{MinDistance} - (\text{RemainingFuel} \times 2.0 \times e^{-\frac{\text{MinDistance}}{500}})\).
* **Stagnation Recovery:** If the rolling median fitness plateaus (Δ < 0.1%) across 4 sequential generations, an **Emergency Repopulation Sweep** overrides the pool, injecting fresh high-variance seeds to escape local optima traps.

---

## ⚙️ Core Aerospace Constants Baseline

| Constant | Value | Description |
| :--- | :--- | :--- |
| `ThrustForce` | `1500.0` | Baseline hardware nominal thrust capability |
| `OneMomentSimulation` | `0.01s` | Euler-Cromer integration time step (Δ t) |
| `GravitationalConstant` | `9.81 m/s²` | Surface gravitational acceleration (g₀) |
| `RadiusPlanetEarth` | `6,371,000.0m`| Mean radius of Earth (\(R_E\)) for gravity decay |
| `AirResistance` | `0.02` | Sea-level baseline aerodynamic drag coefficient (\(C_d\)) |
| `FailedLaunchPenalty` | `50000.0` | Fitness penalty added for rockets failing to clear 5m apogee |

---

## 🚀 Execution & Deployment Steps

### 1. Initialize the Telemetry Dashboard
Boot up your Python visualization environment to bind the socket interface and await the Go engine handshake:
```bash
python radar.py
```

### 2. Fire Up the Evolutionary Engine
Open a separate terminal shell, navigate to the module directory, and launch the compiler runtime:
```bash
cd rocketBehavior
go run .
```
> ⚠️ **Critical Execution Note:** The trailing dot (`.`) is mandatory. It instructs the Go compiler toolchain to discover and bind the package structures of the internal directories (`simulation` and `genetic`).

### 3. Interactive CLI Configuration
Once initialized, provide your execution variables via the step-by-step terminal prompts:
1. **Target:** Comma-separated coordinates in meters: `X, Y, Z` (e.g., `12000, 0, 10`).
2. **Core Rocket DNA:** Initial seed variables: `Pitch(deg), Yaw(deg), Fuel(kg), BurnRate(kg/s), BodyMass(kg)`.
3. **Constraints:** GA population layouts: `PopulationSize, TotalGenerations` (e.g., `500, 100`).

The simulation will stream telemetry chunks to the visual canvas and automatically exit once a specimen safely hits the **Adaptive Threshold** (\(\max(80.0, \text{InitialMedian} \times 0.005)\)).
