# The Algorithm

When reaching the final level, the game enters the random phase: each level is generated randomly and the number of
shovels needed are computed according to the generated game.

## The problem

The game has a NxN grid with cells that could be in one of 
this state

- empty
- digged
- planted

The user can dig an empty cell, but can't dig a planted cell.

Once a cell is digged, its Chebyshev neighbors (their 3x3 neighbors) are marked as "watered".

A Plant is wateres if its cell is watered, meaning there is at least one of its neighbors digged.

The goal of the algorithm is to find the minimum number of cell that need to be digged to water all the plants.

## Modeling and Semplifications

A preliminar model can be drawn by imagining each cells as
nodes in a graph. The nodes are connected vertically and horizontally by edges with specific weights. Weights to 1 if the cells are not digged. Weights to 0 if they are aready digged.

We can simplify the problem by removing the Chebyshev neigbhors logic and creating a virtual target for each plant.
The neighboring cells would have weight = 0 connected to the virtual target.

The problem is now reduced to a `Group Steiner Tree problem`

## Steiner Tree Problem


## Group Steiner Tree problem