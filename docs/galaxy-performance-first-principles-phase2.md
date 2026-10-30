# Galaxy Graph First-Principles Architecture & Redesign (Phase 2)

## 1. Overview & Architectural Philosophy

The Phase 1 investigation demonstrated that Sigma v4's WebGL renderer and modern GPUs can effortlessly render our 3,905 nodes and 21,650 edges in **0.2 milliseconds at over 100 FPS**.

The performance bottleneck is not rendering; it is **software architecture**:
- Running heavy application logic (filtering, label culling, 2D canvas draws) per animation frame on the browser's single UI thread.
- Defeating Sigma's incremental rendering pipeline by using generic JS reducers that trigger global $O(V + E)$ recomputations.
- Artificial debounces and synchronous scans on hover events.

When designed **from first principles**, the graph architecture adheres to three core rules:
1. **The Animation Frame is Sacred**: Never allocate memory or iterate the graph during pan or zoom frames. The GPU already has the data; let it render undisturbed.
2. **Lean State Updates, Not Full Reducers**: Leverage Sigma v4's native state declarations (`setNodeState`, `setNodesState`) so hover/selection changes only update dirty vertices/edges without invalidating clean ones.
3. **Labels Belong in WebGL**: Offload hub and constellation labels to Sigma v4's native WebGL SDF label renderer, restricting 2D Canvas work to a single, targeted focal pill rendered on-demand.

---

## 2. Deep Dive: First-Principles Redesign

### Area 1: Label Rendering — Moving to WebGL SDF & Targeted Tooltip

#### Current Anti-Pattern:
`drawOverlay()` loops all 3,905 nodes on every single frame, calls `getNodeDisplayData` (generating ~240,000 garbage-collected objects/second), runs spatial culling in JS, and paints text onto a 2D canvas using `strokeText` and `fillText`.

#### First-Principles Approach:
1. **Sigma v4 Native SDF Labels for Static/Constellation Hubs**:
   - Sigma v4 includes a built-in Signed Distance Field (SDF) WebGL label renderer with GPU-accelerated spatial grid culling (`labelGridCellSize`, `labelRenderedSizeThreshold`).
   - By enabling native WebGL labels for major genre hubs and artists above a size threshold, label drawing is 100% GPU-accelerated. Panning and zooming cost 0 ms on the CPU.
2. **A Single-Element HTML / SVG / 2D Overlay for the Focused Node**:
   - The dark pill with the kind glyph (🎨 / 🎵), rank, and border only ever exists for **at most two nodes** at any moment: the `selectedNode` and the `hoveredNode`.
   - Instead of clearing and redrawing an entire canvas for 3,905 items, update an absolutely-positioned lightweight DOM/SVG element (or draw exactly 1 or 2 pills to the canvas) **only** when the hover/selection state actually changes.

---

### Area 2: State Management & Eliminating Reducers

#### Current Anti-Pattern:
`nodeReducer` and `edgeReducer` are passed as functions to `new Sigma(...)`.
- In Sigma v4, passing a reducer forces `dependency = "graph-state"`.
- Every gesture, mouse movement, or wheel event that alters `isZooming` or `isPanning` causes Sigma to re-evaluate the reducers across all 3,905 nodes and 21,650 edges.
- Calling `sigma.scheduleRefresh()` recalculates every single edge attribute in JS.

#### First-Principles Approach:
Sigma v4 was redesigned specifically to replace v3-style reducers with a reactive state manager:
1. **Declare Styles Declaratively**:
   ```ts
   styles: {
     nodes: [
       {
         size: { attribute: "size" },
         color: { attribute: "color" },
         // Declarative state-driven styling:
         opacity: {
           rules: [
             { when: { isHighlighted: true }, value: 1.0 },
             { when: { isHovered: true }, value: 1.0 },
             { when: { isDimmed: true }, value: 0.15 }
           ],
           defaultValue: 1.0
         }
       }
     ],
     edges: [
       {
         color: { attribute: "color" },
         opacity: {
           rules: [
             { when: { isHighlighted: true }, value: 0.8 },
             { when: { isDimmed: true }, value: 0.02 }
           ],
           defaultValue: { attribute: "opacity" }
         }
       }
     ]
   }
   ```
2. **Patch Only Dirty Elements on Hover**:
   When hovering node $U$:
   - Do not call `scheduleRefresh()`.
   - Call `sigma.setNodeState(U, { isHovered: true })` and `sigma.setNodesState(neighbors, { isHighlighted: true })`.
   - Sigma updates **only the buffers of the modified items** using `gl.bufferSubData`.
   - Clean nodes and non-incident edges are completely untouched. Frame budget impact: **< 1 ms** instead of **134 ms**.

---

### Area 3: Instant Hover Responsiveness (< 16 ms)

#### Current Anti-Pattern:
- 150 ms debounce via `setTimeout` in `enterNode`.
- Synchronous neighbor scan across all incident edges.

#### First-Principles Approach:
1. **Remove the 150 ms Debounce**:
   With targeted state patching, state updates take < 1 ms. Debouncing is not only unnecessary, it artificially creates the perception of sluggishness. Hover should react on the very next animation frame (0–16 ms).
2. **Pre-Index Neighborhoods**:
   Instead of filtering incident edges on the fly during `enterNode`:
   - Pre-compute an adjacency map on the Graphology instance or in a fast typed array lookup (`neighborMap.get(nodeId)`).
   - In `enterNode`:
     ```ts
     const neighbors = precomputedAdjacency.get(nodeId);
     sigma.setNodeState(nodeId, { isHovered: true });
     sigma.setNodesState(neighbors, { isHighlighted: true });
     ```

---

### Area 4: Filtering (Artist Toggle & 14-Day Window)

#### Current Anti-Pattern:
`nodeReducer` and `edgeReducer` inspect filter sets on every frame to decide if `{ visibility: "hidden" }`.

#### First-Principles Approach:
Graphology & Sigma best practices (per [Graphology Performance Tips](https://graphology.github.io/performance-tips.html)):
1. **Active Filter Subgraph vs. Reducer Visibility**:
   - For major toggles (like "Show Artists" or changing the 14-day sliding time window), modifying edge visibility through continuous reducers wastes CPU cycles.
   - **Alternative A (Batch Attribute Updates)**: When the window changes, iterate the window diff (nodes entering or leaving the window) and update their `hidden` or `visibility` attribute once in Graphology.
   - **Alternative B (Dedicated Subgraph View)**: Maintain the master dataset in a base Graphology graph and project active items into the rendered Sigma graph using `graphology-subgraph` or fast add/drop passes. Since Sigma v4 re-indexes GPU buffers rapidly when structural changes are batched, updating the graph explicitly on filter change leaves all pan/zoom/hover paths completely unburdened by filter logic.

---

### Area 5: Graph Layout & Backbone Optimization

#### Current Status:
The macro-constellation layout (Louvain + separated constellations + Procrustes rotation) produces a visually distinct map, but runs synchronously on the main thread during component mount:
- Louvain clustering
- 48 local FA2 iterations per constellation
- 220 iterations of affinity relaxation
- Circle collision separation (`separateCircles`)

#### First-Principles Improvements:
1. **Web Worker Layout Execution**:
   - Graph topology, Louvain clustering, and ForceAtlas2 layout algorithms in Graphology are pure data calculations with zero DOM/WebGL dependencies.
   - Running them inside a Web Worker prevents initial UI blocking and allows smooth loading indicators.
2. **Server-Side / Build-Time Layout Caching**:
   - Vaultbot's graph data only updates when the background Spotify polling job archives tracks.
   - Rather than having every client recalculate Louvain communities and ForceAtlas2 coordinates from scratch on page load, pre-calculate or cache layout coordinates $(x, y)$ and community IDs. The client then simply ingests coordinates and mounts immediately.

---

## 3. Comparison Matrix: Current vs. Proposed

| Feature | Current Implementation | First-Principles Redesign |
|---|---|---|
| **Label Rendering** | 2D Canvas looping 3,905 nodes/frame; calls `getNodeDisplayData` allocating thousands of objects. | Sigma v4 native WebGL SDF labels for hubs; single-node DOM/SVG overlay for active hover pill. |
| **Node/Edge Styling** | Arbitrary `nodeReducer` & `edgeReducer` forcing global `"graph-state"` invalidations. | Declarative Sigma v4 style rules with reactive state manager (`setNodeState`, `setNodesState`). |
| **Hover Latency** | 170 ms – 480 ms (150 ms debounce + full graph re-evaluation). | < 16 ms (Zero debounce, pre-computed adjacency lookup, lean GPU buffer sub-data updates). |
| **Pan / Zoom FPS** | 20 – 35 FPS (CPU-bound by 2D canvas & state checks). | 60 – 120 FPS rock-steady (GPU handles transforms natively; CPU idle). |
| **Layout Execution** | Synchronous on main UI thread on mount. | Web Worker or precomputed/cached coordinates. |

---

## 4. Proposed Phased Implementation Plan

1. **Phase 2.1: State Management & Reducer Removal**
   - Replace `nodeReducer` and `edgeReducer` with Sigma v4 declarative style rules and `setNodeState`.
   - Eliminate full `scheduleRefresh()` on hover/zoom.
2. **Phase 2.2: Native SDF WebGL Labels & Targeted Hover Pill**
   - Switch genre hub labels to Sigma v4 native SDF labels with grid culling.
   - Simplify overlay canvas / DOM element to only draw the currently hovered/selected node badge.
3. **Phase 2.3: Hover Responsiveness**
   - Remove the 150 ms debounce timer.
   - Pre-index node neighbor sets so hover highlights are instantaneous.
4. **Phase 2.4: Layout & Window Filtering Optimization**
   - Clean up time-window filtering to batch update visibility rather than checking sets per reducer execution.
   - Offload or cache layout positioning.
