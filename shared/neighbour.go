package shared

type Neighbour struct {
	Addr                string
	NID                 int
	NMap                Map
	NeighbourCoordinate Coordinate
	IsWithinCR			bool
}
