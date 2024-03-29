package shared

import "math"
// TODO - comment: HERE ::: delete this struct
type Path struct {
	ListOfPCoordinates []PointStruct
}

// ----------------------------------------- FUNCTIONS ---------------------------------------------------------- //
func CreatePathBetweenTwoPoints(sp Coordinate, dp Coordinate) Path {
	var myPath []PointStruct
	delX := Round(dp.X - sp.X)
	delY := Round(dp.Y - sp.Y)
	//iteration := int(math.Abs(delX) + math.Abs(delY))

	//create the path in X direction
	for i := 0; i < int(math.Abs(delX)); i++ {
		if delX > 0 {
			myPath = append(myPath, PointStruct{Point: Coordinate{1, 0}})
		} else if delX < 0 {
			myPath = append(myPath, PointStruct{Point: Coordinate{-1, 0}})
		} else {
			//do nonthing since the delX is 0
		}
	}

	//create path in Y direction
	for i := 0; i < int(math.Abs(delY)); i++ {
		if delY > 0 {
			myPath = append(myPath, PointStruct{Point: Coordinate{0, 1}})
		} else if delY < 0 {
			myPath = append(myPath, PointStruct{Point: Coordinate{0, -1}})
		} else {
			//do nonthing since the delY is 0
		}
	}

	return Path{myPath}
}
