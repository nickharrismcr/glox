class Cube

init (...) set position extent and colour

add_to_batch ( batch )  add to cube draw batch


class CompositeCube

init(pos) 

create a black cube with coloured windows in the front/back/side faces, slightly protruding to prevent Z clash
at given pos, using a set of Cubes

add_to_batch (batch )

add all cubes in the composite to the given batch uding Cube.add_to_batch


class Stack 
init (pos,height)

create a number of CompositeCubes of varying sizes and hold their positions in a stack so they abut correctly, 
at a given grid position

add_to_batch(batch)
add all the composite cubes to the batch using the relevant method

class Grid

init (size)

create a size x size grid of points holding stack/street positions. stacks can only be on positions where x and y are odd so there are channels between them. each stack created give it a random height up to a const max

00123456789
1x.x.x.x.x
2.........
3x.x.x.x.x
4.........

class Controller

controls a camera. controller can be either starting, moving, stopping, or rotating.
it can face n s e w  ( consts for direction ) so holds current direction and speed

if it is starting, its speed will ramp from zero to a const maximum over 60 ticks and it will change state to moving
if it is moving then its target will be the point on the grid in front of it. 
when the controller reaches the next point it will decide to keep moving or transition to stopping in which case its speed will ramp down over 60 ticks until it is stopped at the next point. it will then transition to rotating and turn 90 degrees left or right, random choice , or 180 degrees if it is at the edge of the grid over 60 ticks. then transition to starting. 

the controller will start in an even grid position in the centre of the grid. 

setup will create a batch,  a grid, a controller, and instruct the grid to create stacks and add them to batches.
there will be a main draw loop which will update the camera position and draw the batch. 