## 📈 Project Status & Release Roadmap

- [x] **v1.0.0 — Monolithic Prototype**
  - [x] High-fidelity physical simulation engine in Go (Gravity gradients & Aerodynamic drag).
  - [x] Integration of the Evolutionary Genetic Algorithm (GA) with 5-gene chromosome DNA.
  - [x] Real-time TCP network bridge linking Go & Python pipelines.
- [x] **v1.1.0 — Architecture Optimization & Bugfix Release**
  - [x] **Modular Monolith Transition:** Separation of concerns by splitting the single file into isolated packages (`simulation`, `genetic`, `network`).
  - [x] **Population Size Guardrail:** Fixed a critical bug in `GeneticCalculation` that caused population overflows and `index out of range` panics.
  - [x] **Yaw Mutation Fix:** Resolved a copy-paste bug in the initial mutation cycle where `FirstRocket.YawDegree` was mutated instead of `YawRand`.
  - [x] **Strict Type Validation:** Standardized Go formatting and code visibility via explicit cross-package imports (`simulation.Rocket`).
- [ ] **v2.0.0 — Future Enhancements**
  - [ ] Implementation of full 3D spatial flight dynamics using the `YawDegree` gene.
  - [ ] Dynamic wind vectors and true coriolis effect calculations.

---

## 🏗️ Architectural Refactoring (v1.1.0)

In version 1.1.0, the codebase underwent a major architectural redesign to enforce the **Single Responsibility Principle** and improve maintainability:

```text
rocketBehavior/
├── go.mod                # Core Go Module Definition
├── main.go               # Application Entry Point & CLI UI Controller
├── simulation/
│   ├── models.go         # Shared Entities (Rocket struct & Constants)
│   └── physics.go        # Euler-Cromer Trajectory Calculation Engine
└── genetic/
    └── algorithm.go      # Evolutionary Crossover, Validation, and Selection
```

### Key Enhancements:
1. **Dynamic Crossover Loop:** The `GeneticCalculation` function was rewritten from a rigid hardcoded loop to a dynamic `len(population) < populationSize` constraint. It now perfectly respects user configurations without memory leaks.
2. **Encapsulated Scope Visibility:** Internal logic is hidden within local packages, preventing accidental global data corruption during concurrent evaluations.

---

## 🚀 How to Run the Ecosystem

1. Initialize the Python telemetry listener pipeline to prepare the 3D visual canvas:
   ```bash
   python radar.py
   ```
2. Navigate to the core project directory:
   ```bash
   cd rocketBehavior
   ```
3. Run the concurrent Go evolutionary engine by compiling the entire root module (the `.` is mandatory for package discovery):
   ```bash
   go run .
   ```
4. Input target spatial coordinates (e.g., `10000, 0, 5`) and design the core DNA parameters of your baseline specimen to watch the algorithm converge across generations.