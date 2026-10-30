# Galaxy Graph Performance Investigation (Phase 1)

## Executive Summary

Following PR #290, the Playlist Galaxy page (`/galaxy`) achieved good initial load times, but interactive responsiveness degraded severely:
- **Panning / dragging**: Severe frame drops (~35 FPS avg, with frequent spikes > 200 ms).
- **Zooming (wheel)**: Stuttering (~20 FPS avg, with spikes up to 450 ms and multiple long tasks).
- **Node hover**: Significant latency (170 ms to 480 ms from mouse arrival to rendered highlight), accompanied by high CPU spikes and UI freeze.

Hardware profiling using Chrome with hardware WebGL2 (Vulkan backend) across our production dataset (**1,101 genres, 2,804 artists [3,905 nodes total], 21,650 edges [12,412 backbone]**) revealed that **the GPU and WebGL renderer are not the bottleneck**. Sigma v4 renders all 21,650 edges and 3,905 nodes in **0.2 milliseconds (< 1 ms)** at over 100 FPS when undisturbed. 

The sluggishness is **100% main thread CPU starvation** caused by three architectural bottlenecks:
1. An expensive 2D HTML5 canvas overlay recalculating all 3,905 nodes every frame via unmemoized `getNodeDisplayData()` calls.
2. Sigma v4 style reducers (`nodeReducer` and `edgeReducer`) that force Sigma's style dependency classification to `"graph-state"`, triggering full graph re-evaluations across 25,500+ items during zoom and hover.
3. An artificial 150 ms debounce in combination with heavy unindexed neighbor filtering on hover.

---

## Benchmark Attribution Data

We isolated and profiled individual subsystems using automated Chrome DevTools Protocol (CDP) sampling sessions:

| Test Scenario | Drag Avg Frame (target < 16.6 ms) | Drag p95 Frame | Zoom Avg Frame | Zoom p95 Frame | Sigma WebGL Render Time | Long Tasks (> 50 ms) |
|---|---|---|---|---|---|---|
| **A. Baseline (Current implementation)** | **28.2 ms (~35 FPS)** | **29.2 ms** | **50.5 ms (~20 FPS)** | **150.2 ms** | 28.9 ms | **25 long tasks** |
| **B. Detach custom 2D Canvas Overlay** | **12.5 ms (~80 FPS)** | **17.7 ms** | **23.6 ms (~42 FPS)** | **103.8 ms** | 4.2 ms | 5 long tasks |
| **C. Detach Overlay + Remove Reducers** | **10.6 ms (~95 FPS)** | **17.5 ms** | **7.6 ms (~130 FPS)** | **17.6 ms** | **0.2 ms** | **0 long tasks** |
| **D. All 21.6k edges drawn every frame** | **9.6 ms (>100 FPS)** | **17.3 ms** | **9.0 ms (>110 FPS)** | **17.6 ms** | **0.2 ms** | **0 long tasks** |

---

## Detailed Root Cause Analysis

### 1. Frame-Dropping During Drag / Pan: The 2D Overlay Canvas

In `web/src/lib/GenreGraph.svelte` (lines 638–734), an HTML5 2D `<canvas>` overlay runs `drawOverlay()` on every single animation frame via `sigma.on("afterRender", drawOverlay)`.

Inside `drawOverlay`:
```ts
g.forEachNode((node, attrs) => {
    const display = sigma.getNodeDisplayData(node);
    if (!display || display.visibility === "hidden") return;
    if (!display.label) return;
    ...
    const { x, y } = sigma.framedGraphToViewport(display);
    ...
});
```

#### Why this starves the CPU:
1. **Unbounded Node Iteration**: Every frame, JS loops over all 3,905 nodes and computes coordinates and display criteria.
2. **Object Allocation Penalty in `getNodeDisplayData`**:
   In Sigma v4 (`sigma.esm.js`):
   ```js
   getNodeDisplayData(key) {
     var node = this.internals.nodeDataCache[key];
     return node ? Object.assign({}, node) : undefined;
   }
   ```
   Each node record has 35 display attributes. Calling this for 3,905 nodes creates **3,905 fresh object allocations on every frame** (~240,000 objects per second during panning).
   - In our CPU profiles, `getNodeDisplayData` alone consumed **2,231 ms of CPU time** in a 1.5-second drag, turning what should be a 0.5 ms frame into a 28 ms frame.

---

### 2. Sluggish Node Hover Latency (170 ms – 480 ms)

When moving the mouse over a node:
1. **Artificial 150 ms Debounce**:
   `enterNode` wraps hover handling in `setTimeout(..., 150)`. This guarantees an immediate 150 ms floor before the app even attempts to respond.
2. **Synchronous Neighbor Edge Scanning**:
   `activeNeighborsOf(g, hoveredNode, ...)` iterates all incident edges of the target node and checks filter sets. For high-degree hub nodes (e.g. `g:45` with degree 435), this creates noticeable CPU spikes.
3. **Reducers Invalidate Dirty-Tracking**:
   `sigmaInst.scheduleRefresh({ skipIndexation: true })` is called when hover state changes.
   Because custom `nodeReducer` and `edgeReducer` functions are registered on the Sigma instance:
   - Sigma cannot perform lean attribute patching.
   - It re-evaluates all 3,905 nodes and all 21,650 edges sequentially on the CPU (`_objectSpread2` alone accounted for **1,777 ms** of CPU time during hover testing).
   - `render()` average duration jumped from 0.2 ms to **134.5 ms** (peaking at **322.7 ms**).
   - Measured latency from pointer arrival to visual highlight rendering ranged from **172 ms to 479 ms**.

---

### 3. Frame Drops During Zoom: Graph State Invalidation

When scrolling the mouse wheel:
1. Sigma's mouse captor toggles internal graph state flags (`isZooming: true`, then back to `isIdle`).
2. Because opaque `nodeReducer` and `edgeReducer` functions are present, Sigma conservatively marks the style dependency as `"graph-state"`.
3. Consequently, Sigma invokes `refreshState()` on every zoom step (**116 ms** average per call, invoked 23 times in 1 second).
4. `drawOverlay()` executes on every frame after each zoom render, compounding the delay.

---

## Current Graph Construction & Layout Architecture

The graph pipeline currently executes in two main stages:

### Stage 1: Data Decoration (`buildMixedGraph` in `mixed-graph.ts`)
1. Ingests vertices and edges from `/api/graph`.
2. Computes log-scaled node sizes and edge weights.
3. Selects top-$k$ backbone edges per node (top 4 for genre-genre, top 3 for genre-artist / artist-artist) via `selectBackboneEdges`.
4. Creates a `backboneGraph` and runs **Louvain community detection** (`graphology-communities-louvain`).
5. Merges micro-communities (< 25 nodes) into neighboring communities sharing the strongest connection.
6. Decorates node colors based on community palette, desaturating artist nodes by 32%.
7. Computes edge styles (opacity, width, colors).

### Stage 2: Spatial Layout (`assignGalaxyLayout` in `GenreGraph.svelte` & `graph.ts`)
1. Partitions graph into connected components ($\ge 10$ nodes = core; $< 10$ nodes = satellites).
2. **Local Community Layout**:
   - For each Louvain community, builds an isolated subgraph with its backbone edges.
   - Initializes positions using `circlepack`.
   - Runs local **ForceAtlas2 (48 iterations)** with strong gravity.
   - Scales positions by star density (RMS normalization).
   - Runs pairwise circle collision separation (`separateCircles`).
3. **Macro Constellation Layout**:
   - Calculates inter-community affinity matrix.
   - Positions macro-constellation centers using an affinity spiral + 220-iteration physics relaxation.
   - Aligns principal axis horizontally (PCA) for widescreen viewports.
   - Applies hard collision separation (`channelGap` = 54).
4. **Orientation & Satellites**:
   - Rotates each constellation via 2D Procrustes alignment so bridge genres face neighboring constellations.
   - Arranges satellite components along an outer elliptical boundary.
