package shared

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"net/rpc"
	"os"
	"time"
	"github.com/DistributedClocks/GoVector/govec"
	"github.com/fatih/set"
	"encoding/json"
)

const XMIN = "xmin"
const XMAX = "xmax"
const YMIN = "ymix"
const YMAX = "ymax"
const EXRADIUS = 6
const TIMETOJOIN = 10*time.Second
const TIMETOJOINSECONDUNIT = 10

type JoiningInfo struct {
	joiningTime  time.Time
	firstTimeJoining bool
}


type RobotLog struct {
	CurrTask    TaskPayload
	RMap        Map
	CurLocation Coordinate
}

type RobotStruct struct {
	CurrTask           TaskPayload
	PossibleNeighbours *set.Set
	RobotID            int // hardcoded
	RobotIP            string
	RobotEnergy        int
	RobotListenConn    *rpc.Client
	//RobotNeighbours    []Neighbour
	RobotNeighbours	   map[int]Neighbour
	RMap               Map
	CurPath            Path
	// CurPath        []Coordinate // TODO: yo micheal here uncomment, n delete the whole struct
	CurLocation   Coordinate    // TODO why isn't type coordinate instead?
	ReceivedTasks []TaskPayload // change this later
	JoiningSig    chan Neighbour
	BusySig       chan bool
	WaitingSig    chan bool
	FreeSpaceSig  chan bool
	WallSig       chan bool
	RightWallSig  chan bool
	LeftWallSig   chan bool
	WalkSig       chan bool
	Logname       string
	Logger        *govec.GoLog
	State         RobotState
    joinInfo      JoiningInfo
}

type Robot interface {
	SendMyMap()
	MergeMaps(neighbourMaps []Map) error
	Explore() error //make a step base on the robat's current path
	GetMap() Map
	SendFreeSpaceSig()
}

var robotStruct RobotStruct

// FN: this robot sends map and ID to its neighbours
func (r *RobotStruct) SendMyMap() {
	return
}

type RobotState int

const (
	ROAM RobotState = iota
	JOIN RobotState = iota
	BUSY RobotState = iota
)

func (r *RobotStruct) SendFreeSpaceSig() {
	fmt.Println("got here")
	r.FreeSpaceSig <- true
}

//error is not nil when the task queue is empty
// FN: Return list of destination points for each node in the network (one point for each node)
//     This robots destination point is placed at the beginning
//		TODO: comment : whats the error for?
func (r *RobotStruct) TaskCreation() ([]PointStruct, error) {

	xmin := r.FindMapExtrema(XMIN)
	xmax := r.FindMapExtrema(XMAX)
	ymin := r.FindMapExtrema(YMIN)
	ymax := r.FindMapExtrema(YMAX)

	center := Coordinate{Round(float64((xmax - xmin) / 2)), Round(float64((ymax - ymin) / 2))}
	center.X = Round(center.X)
	center.Y = Round(center.Y)

	DestNum := len(r.RobotNeighbours) + 1
	//fmt.Println("DESTNum is ")
	//fmt.Println(DestNum)
	//fmt.Println(r.RobotNeighbours)


	DestPoints := FindDestPoints(DestNum, center)

	// move DestpointForMe to beginning of list
	DestPointForMe := r.FindClosestDest(DestPoints)

	tempEle := DestPoints[0]
	for idx, value := range DestPoints {
		if value == DestPointForMe {
			DestPoints[0] = value
			DestPoints[idx] = tempEle
			break
		}
	}

	return DestPoints, nil

}

// TODO: comment: yo why isnt this a switch statement?
func (r *RobotStruct) FindMapExtrema(e string) float64 {

	if e == XMAX {
		var xMax float64 = math.MinInt64
		for _, point := range r.RMap.ExploredPath {
			if xMax < point.Point.X {
				xMax = point.Point.X
			}
		}

		if len(r.RMap.ExploredPath) == 0{
			return 0.0
		}

		return Round(xMax)
	} else if e == XMIN {
		var xMin float64 = math.MaxFloat64
		for _, point := range r.RMap.ExploredPath {
			if xMin > point.Point.X {
				xMin = point.Point.X
			}
		}

		if len(r.RMap.ExploredPath) == 0{
			return 0.0
		}
		return Round(xMin)
	} else if e == YMAX {
		var yMax float64 = math.MinInt64
		for _, point := range r.RMap.ExploredPath {
			if yMax < point.Point.Y {
				yMax = point.Point.Y
			}
		}

		if len(r.RMap.ExploredPath) == 0{
			return 0.0
		}
		return Round(yMax)
	} else {
		var yMin float64 = math.MaxFloat64
		for _, point := range r.RMap.ExploredPath {
			if yMin > point.Point.Y {
				yMin = point.Point.Y
			}
		}

		if len(r.RMap.ExploredPath) == 0{
			return 0.0
		}
		return Round(yMin)
	}
}

// FN: Find destination point that will require the least amound of energy to go to
func (r *RobotStruct) FindClosestDest(lodp []PointStruct) PointStruct {
	dist := math.MaxFloat64
	var rdp PointStruct
	for _, dp := range lodp {
		del := DistBtwnTwoPoints(dp.Point, r.CurLocation)
		if del < dist {
			dist = del
			rdp = dp
		}
	}

	return rdp
}

func (r *RobotStruct) RespondToButtons() error {
	// This function listen to GPIO
	for {
		fmt.Println(" Press j to send JoinSig \n Press b to send BusySig \n Press w to send WaitSig \n Press s to send WalkSig \n Press o to send WallSig")
		buf := bufio.NewReader(os.Stdin)
		signal, err := buf.ReadByte()
		if err != nil {
			fmt.Println(err)
		}
		command := string(signal)

		if command == "j" {

			r.JoiningSig <- Neighbour{
				Addr: ":8080",
				NID: 1,
				NMap: RandomMapGenerator(),
				NeighbourCoordinate: Coordinate{4.0, 5.0},
			}

		} else if command == "b" {
			r.BusySig <- true
		} else if command == "w" {
			r.WaitingSig <- true
		} else if command == "s" {
			r.FreeSpaceSig <- true
		} else if command == "o" {
			r.WallSig <- true
		}
	}
}

func (r *RobotStruct) Explore() error {
	for {
		fmt.Println("Explore() at the top")
		if len(r.CurPath.ListOfPCoordinates) == 0 {
			dpts, err := r.TaskCreation()
			if err != nil {
				fmt.Println("error generating task")
			}
			var newPath Path
			// TESTING
			newPath = CreatePathBetweenTwoPoints(r.CurLocation, dpts[0].Point)
			// TESTING

			/**** SHOULDN'T BE COMMENTED OUT JUST FOR TESTING
			if len(dpts) == 1 {
				newPath = CreatePathBetweenTwoPoints(r.CurLocation, dpts[0].Point)
			} else {
				fmt.Println("Explore() > 1 destination point returned when it should have no neighbours")
				return CodeError("Explore() > 1 destination point returned when it should have no neighbours")
			}
			*/
			r.CurPath = newPath
			// DISPLAY task with GPIO
		}

		fmt.Println("\nWaiting for signal to proceed.....")

		select {
		case <-r.FreeSpaceSig:
			fmt.Println("FreeSpaceSig received")
			r.UpdateMap(FreeSpace)
			r.UpdateCurLocation()
			r.UpdatePath()
			// r.SetCurrentLocation()
			// r.TookOneStep() //remove the first element from r.CurPath.ListOfPCoordinates

			// Display task with GPIO
		case <-r.WallSig:
			r.UpdateMap(Wall)
			r.ModifyPathForWall()
			// Display task with GPIO
		case <-r.RightWallSig:
			r.UpdateMap(RightWall)
		case <-r.LeftWallSig:
			r.UpdateMap(LeftWall)
		case <-r.BusySig: // TODO whole thing
			fmt.Println("busy sig received")

			listOfNeighbourMaps :=  make([]Map, len(r.RobotNeighbours))

			fmt.Println("Getting the maps from the neighbour.................")
			for _, nei := range r.RobotNeighbours {
				neighbourMap := Map{}
				client, err := rpc.Dial("tcp", nei.Addr)
				if err != nil {
					fmt.Println("Error in connecting with neighbour")
					fmt.Println(err)
					continue
				}

				err = client.Call("RobotRPC.ReceiveMap", false, &neighbourMap)

				//Logging
				rawMap, _ := json.MarshalIndent(neighbourMap, "", "")
				fmt.Printf("Receive map from %s \n", nei.Addr)
				fmt.Println(string(rawMap))


				if err != nil {
					fmt.Println("Error in getting the neighbour's map")
					fmt.Println(err)
					continue
				}
				listOfNeighbourMaps = append(listOfNeighbourMaps, neighbourMap)
			}

			fmt.Println()
			fmt.Println("Retrieved the map. Start merging..........")
			fmt.Println()

			r.MergeMaps(listOfNeighbourMaps)
			fmt.Println("Finished Merging")

			// Exchange my map with neighbours
			// Wait till maps from all neighbours are recevied
			// Merge my map with neighbours
			// Create tasks for current robot network
			tasks, _ := r.TaskCreation()
			// Allocate tasks to current robot network
			r.CurPath = CreatePathBetweenTwoPoints(r.CurLocation, tasks[0].Point)
			// r.CurrTask = tasks[0]
			fmt.Println("tasks length is")
			fmt.Println(len(tasks))

			fmt.Println("number of neighbour is ")
			fmt.Println(len(r.RobotNeighbours))
			r.TaskAllocationToNeighbours(tasks[1:])

			// Wait for tasks from each neighbour
			r.WaitForEnoughTaskFromNeighbours()
			taskToDo := r.PickTaskWithLowestID()
			// r.CurrTask = taskToDo
			if taskToDo.SenderID < r.RobotID {
				r.CurPath = CreatePathBetweenTwoPoints(r.CurLocation, taskToDo.DestPoint.Point)
			}
			r.RespondToNeighoursAboutTask(taskToDo)
			// Respond to each task given by my fellow robots
			//       r.decideTaskTodo()
			// Agree with everyone in the network of who assigned the task
			//		- YES --> set newTaskthreshold thing, create new path based on new task
			//		- NO --> handle case ?
			// set busysig off
			// procede with new task
			fmt.Println("THE CURRENT MAP IS")
			fmt.Println(r.RMap)
		case <-r.WaitingSig: // TODO
			// keep pinging the neighbour that is within it's communication radius
			// if neighbour in busy state
			// YES -> keep pinging
			// NO -> - turn WaitingSig off
			//		 - turn JoingingSig on
		}
	}
}

func (r *RobotStruct) ModifyPathForWall() {

	wallCoor := r.CurPath.ListOfPCoordinates[0]
	tempList := r.CurPath.ListOfPCoordinates
	//tempList := make([]Coordinate, 0)
	for i, c := range tempList {
		if wallCoor == c {
			continue
		}
		r.CurPath.ListOfPCoordinates = r.CurPath.ListOfPCoordinates[i:]
		break
	}
}

func (r *RobotStruct) TookOneStep() {
	r.CurPath.ListOfPCoordinates = r.CurPath.ListOfPCoordinates[1:]
}

// FN: Removes the just traversed coordinate (first element in the Path list)
func (r *RobotStruct) UpdatePath() {
	r.CurPath.ListOfPCoordinates = r.CurPath.ListOfPCoordinates[1:]
}

//update explored point in map:
// pointkind: 1 - freespace
// 			  2 - wall at current coordinate
// 			  3 - right bumper wall
// 			  4 - left bumper wall
func (r *RobotStruct) UpdateMap(b Button) error {

	var justExploredPoint PointStruct

	switch b {
	case FreeSpace:
		{
			justExploredPoint.Point.X = r.CurLocation.X + r.CurPath.ListOfPCoordinates[0].Point.X
			justExploredPoint.Point.Y = r.CurLocation.Y + r.CurPath.ListOfPCoordinates[0].Point.Y
			justExploredPoint.PointKind = true
			justExploredPoint.Traversed = true
			justExploredPoint.TraversedTime = time.Now().Unix()

			break
		}
	case Wall:
		{
			justExploredPoint.Point.X = r.CurLocation.X + r.CurPath.ListOfPCoordinates[0].Point.X
			justExploredPoint.Point.Y = r.CurLocation.Y + r.CurPath.ListOfPCoordinates[0].Point.Y
			justExploredPoint.PointKind = false
			justExploredPoint.Traversed = true
			justExploredPoint.TraversedTime = time.Now().Unix()

			break
		}
	case RightWall:
		{
			justExploredPoint.Point.X = r.CurLocation.X + r.CurPath.ListOfPCoordinates[0].Point.X + 1
			justExploredPoint.Point.Y = r.CurLocation.Y + r.CurPath.ListOfPCoordinates[0].Point.Y
			justExploredPoint.PointKind = false
			justExploredPoint.Traversed = true
			justExploredPoint.TraversedTime = time.Now().Unix()

			break
		}
	case LeftWall:
		{
			justExploredPoint.Point.X = r.CurLocation.X + r.CurPath.ListOfPCoordinates[0].Point.X - 1
			justExploredPoint.Point.Y = r.CurLocation.Y + r.CurPath.ListOfPCoordinates[0].Point.Y
			justExploredPoint.PointKind = true
			justExploredPoint.Traversed = true
			justExploredPoint.TraversedTime = time.Now().Unix()

			break
		}
	default:
		fmt.Println("UpdateMap () Found incorrect type of wall -- CODE INCORRECT")
		return CodeError("UpdateMap () Found incorrect type of wall")

	}

	oldcoor, exists := r.RMap.ExploredPath[justExploredPoint.Point]
	if exists {
		oldcoor.TraversedTime = justExploredPoint.TraversedTime
		oldcoor.Traversed = justExploredPoint.Traversed
		oldcoor.PointKind = justExploredPoint.PointKind
	}

	return nil
}

func (r *RobotStruct) RespondToNeighoursAboutTask(taskToDo TaskPayload) {
	for _, neighbour := range r.RobotNeighbours {
		client, err := rpc.Dial("tcp", neighbour.Addr)
		if err != nil {
			fmt.Println("There is a problem respoing to neighbour about its task")
		}
		responsePayload := ResponseForNeighbourPayload{}

		if neighbour.Addr == taskToDo.SenderAddr {
			messagepayload := 1
			finalsend := r.Logger.PrepareSend("Sending Message - "+"Accpeting task from my neighbour:"+neighbour.Addr, messagepayload)
			taskResponsePayloadYes := TaskDescisionPayload{
				SenderID:       r.RobotID,
				SenderAddr:     r.RobotIP,
				Descision:      true,
				SendlogMessage: finalsend,
			}
			client.Call("RobotRPC.ReceivePossibleNeighboursPayload", taskResponsePayloadYes, &responsePayload)
		} else {
			messagepayload := 1
			finalsend := r.Logger.PrepareSend("Sending Message - "+"Denying task from my neighbour:"+neighbour.Addr, messagepayload)
			taskResponsePayloadNo := TaskDescisionPayload{
				SenderID:       r.RobotID,
				SenderAddr:     r.RobotIP,
				Descision:      false,
				SendlogMessage: finalsend,
			}
			client.Call("RobotRPC.ReceivePossibleNeighboursPayload", taskResponsePayloadNo, &responsePayload)
		}
	}
}

// Assuming same coordinate system, and each robot has difference ExploredPath
func (r *RobotStruct) MergeMaps(neighbourMaps []Map)  {
	refToOriginalMap := r.RMap

	for _, neighbourRobotMap := range neighbourMaps {

		if len(refToOriginalMap.ExploredPath) == 0 {
			r.RMap.ExploredPath = neighbourRobotMap.ExploredPath
		} else {
			neighbourExploredPath := neighbourRobotMap.ExploredPath

			for neighbourCoordinate, neighbourPointInfo := range neighbourExploredPath {
				if currentPointInfo, ok := r.RMap.ExploredPath[neighbourCoordinate]; ok &&
					currentPointInfo.TraversedTime < neighbourPointInfo.TraversedTime {

					r.RMap.ExploredPath[neighbourCoordinate] = neighbourPointInfo
					continue
				}
				r.RMap.ExploredPath[neighbourCoordinate] = neighbourPointInfo
			}

		}
	}
}

func (r *RobotStruct) GetMap() Map {
	return r.RMap
}

// TODO comment: we dont need this
func (r *RobotStruct) SetCurrentLocation() {
	r.CurLocation = r.CurPath.ListOfPCoordinates[0].Point
}

// TODO comment: update this when path type is updated
func (r *RobotStruct) UpdateCurLocation() {
	r.CurLocation.X = r.CurLocation.X + r.CurPath.ListOfPCoordinates[0].Point.X
	r.CurLocation.Y = r.CurLocation.Y + r.CurPath.ListOfPCoordinates[0].Point.Y
}

func (r *RobotStruct) WaitForEnoughTaskFromNeighbours() {
WaitingForEnoughTask:
	for {
		if len(r.ReceivedTasks) == len(r.RobotNeighbours) {
			fmt.Println("waiting for my neighbours to send me tasks")
			// choose task
			// r.CurPath = something
			// should enter default Roaming state, aka don't need to do anything
			break WaitingForEnoughTask
		}
	}
}

func (r *RobotStruct) PickTaskWithLowestID() TaskPayload {
	localMin := -1
	var taskToDo TaskPayload
	for _, task := range r.ReceivedTasks {
		if task.SenderID < localMin {
			localMin = task.SenderID
			taskToDo = task
		}
	}
	return taskToDo
}

func (r *RobotStruct) TaskAllocationToNeighbours(ldp []PointStruct) {
	//fmt.Printf( "The length of LDPN is  %v \n", len(ldp))
	// TODO: What happens when len(ldp) == 1
	ldpn := ldp
	rand.Seed(time.Now().UnixNano())
	for _, robotNeighbour := range r.RobotNeighbours {
		//fmt.Printf( "The length of LDPN is  %v \n", len(ldpn))
		dpn := ldpn[rand.Intn(len(ldpn))]
		removeElFromlist(dpn, &ldpn)
		fmt.Printf("Current Neighour %s \n", robotNeighbour)
		// fmt.Println(neighbourRoboAddr)
		messagepayload := 1
		finalsend := r.Logger.PrepareSend("Sending Message to Robot"+robotNeighbour.Addr, messagepayload)
		task := &TaskPayload{
			SenderID:       r.RobotID,
			SenderAddr:     r.RobotIP,
			DestPoint:      dpn,
			SendlogMessage: finalsend,
		}

		fmt.Println()
		fmt.Printf("TaskAllocateToNeighbours(%s -------> %s) \n", task.SenderAddr, robotNeighbour.Addr)
		data, _ := json.MarshalIndent(task, "", "")
		fmt.Println(string(data))
		fmt.Println()

		// TESTING UNCOMMENT
		neighbourClient, err := rpc.Dial("tcp", robotNeighbour.Addr)
		if err != nil {
			fmt.Println(err)
		}
		//fmt.Printf("%+v", neighbourClient)
		alive := false
		// Here I send my robot the task
		err = neighbourClient.Call("RobotRPC.ReceiveTask", task, &alive)
		if err != nil {
			fmt.Println(err)
		}
		// TESTING UNCOMMENT
	}
}


// Client -> R2
func (r *RobotStruct) CallNeighbours() {
	for {
		for _, possibleNeighbour := range r.PossibleNeighbours.List() {
			client, err := rpc.Dial("tcp", possibleNeighbour.(string))
			if err != nil {
				r.PossibleNeighbours.Remove(possibleNeighbour)
				continue
			}
			// fmt.Println(client)
			// messagepayload := []byte("Receiving coorindates info from neighbour: " + strconv.Itoa(r.RobotID))
			// finalsend := r.Logger.PrepareSend("Sending Message", messagepayload)
			messagepayload := 1
			finalsend := r.Logger.PrepareSend("Sending Message - "+"Trying to call my neighbour:"+possibleNeighbour.(string), &messagepayload)


			farNeighbourPayload := FarNeighbourPayload{
				NeighbourID:         r.RobotID,
				NeighbourIPAddr:     r.RobotIP,
				NeighbourCoordinate: r.CurLocation,
				NeighbourMap:        r.RMap,
				SendlogMessage:      finalsend,
				State: 				 r.State,
				//ItsNeighbours:       r.RobotNeighbours,
			}

			for _, robot := range r.RobotNeighbours{
				farNeighbourPayload.ItsNeighbours = append(farNeighbourPayload.ItsNeighbours, robot)
			}

			if err != nil {
				fmt.Println(err)
			}
			responsePayload := ResponseForNeighbourPayload{}

			if (r.State == ROAM) {

				client.Call("RobotRPC.ReceivePossibleNeighboursPayload", farNeighbourPayload, &responsePayload)

				//if other robot is in join/roam and within cr, current robot tries joining
				if responsePayload.WithInComRadius && (responsePayload.NeighbourState == JOIN ||  responsePayload.NeighbourState == ROAM) {

					//store everyNode in the network
					//r.RobotNeighbours = append(r.RobotNeighbours, responsePayload.NeighboursNeighbourRobots...)
					for i:= 0; i<len(responsePayload.NeighboursNeighbourRobots); i++{
						if responsePayload.NeighboursNeighbourRobots[i].NID != r.RobotID{
							r.RobotNeighbours[responsePayload.NeighboursNeighbourRobots[i].NID] = responsePayload.NeighboursNeighbourRobots[i]
						}
					}
					r.State = JOIN

				}

				r.joinInfo.joiningTime = time.Now()
				ticker := time.NewTicker(1000 * time.Millisecond)

				//start the time, which time depending on the neighbour's response
				if responsePayload.NeighbourState == JOIN{
					fmt.Println("The Pi has the remaining time of ", responsePayload.RemainingTime)

					//robot start its time using the remainingTime
					go func(){
						counter := 0
						for t := range ticker.C {
							counter += 1
							fmt.Println("Tick at", t)
							fmt.Printf("Counter is %s\n", counter)
							if time.Now().Sub(r.joinInfo.joiningTime) >= (TIMETOJOIN - responsePayload.RemainingTime){
								fmt.Println("WE ARE FINISHED.FUCK 416 -- JOIN")
								ticker.Stop()
								r.State = BUSY
								r.BusySig <- true
							}
						}
					}()


				}else if(responsePayload.NeighbourState == ROAM){

					//robot starts its owner timer
					go func() {
						counter :=0
						for t := range ticker.C {
							counter += 1
							fmt.Println("Tick at", t)
							fmt.Printf("Counter is %s\n", counter)
							if counter >= TIMETOJOINSECONDUNIT{
								fmt.Println("WE ARE FINISHED.FUCK 416 -- ROAM")

								ticker.Stop()
								r.State = BUSY
								r.BusySig <- true

							}
						}
					}()

				}else{
					//do nothing
					//neightbours are in hte busy state
				}
			}

		}
		//time.Sleep(500 * time.Millisecond)
	}
}

// TODO
// Decide the appropriate task that the neighbours assigned it and send response to neighbours
func (r *RobotStruct) decideTaskTodo() {
	// call ReceiveTaskDecsionResponse() here

}

func InitRobot(rID int, initMap Map, logger *govec.GoLog, robotIPAddr string, logname string) *RobotStruct {
	newRobot := RobotStruct{

		PossibleNeighbours: set.New(),
		RobotID:            rID,
		RobotIP:            robotIPAddr,
		RobotNeighbours:    make(map[int]Neighbour),
		RMap:               initMap,
		JoiningSig:         make(chan Neighbour),
		BusySig:            make(chan bool),
		WaitingSig:         make(chan bool),
		FreeSpaceSig:       make(chan bool),
		WallSig:            make(chan bool),
		RightWallSig:       make(chan bool),
		LeftWallSig:        make(chan bool),
		WalkSig:            make(chan bool),
		Logname:            logname,
		Logger:             logger,
		State:              ROAM,
		joinInfo:           JoiningInfo{time.Now(), true},
	}
	// newRobot.CurPath.ListOfPCoordinates = append(newRobot.CurPath.ListOfPCoordinates, shared.PointStruct{PointKind: true})
	return &newRobot
}

func (r *RobotStruct) EncodeRobotLogInfo(robotLog RobotLog) string {
	buf := new(bytes.Buffer)
	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(robotLog)
	if err != nil {
		panic(err)
	}
	output := string(buf.Bytes())
	return output
	// fmt.Println(buf.Bytes())
}

func (r *RobotStruct) ReadFromLog() {
	robotLogContent, _ := ioutil.ReadFile("./" + r.Logname)
	buf := bytes.NewBuffer(robotLogContent)

	var decodedRobotLog RobotLog

	decoder := gob.NewDecoder(buf)
	err := decoder.Decode(&decodedRobotLog)
	if err != nil {
		panic(err)
	}

	r.RMap = decodedRobotLog.RMap
	r.CurLocation = decodedRobotLog.CurLocation
	r.CurrTask = decodedRobotLog.CurrTask
	fmt.Println(decodedRobotLog.RMap)
	fmt.Println(decodedRobotLog.CurLocation)
	fmt.Println(decodedRobotLog.CurrTask)
	fmt.Println("finshed loading from log")
}

func (r *RobotStruct) CreateLog() (*os.File, error) {
	file, err := os.Create("./" + r.Logname)
	if err != nil {
		fmt.Println("error creating robot log")
	}
	return file, err
}

func (r *RobotStruct) ProduceLogInfo() RobotLog {
	robotLog := RobotLog{
		CurrTask:    r.CurrTask,
		RMap:        r.RMap,
		CurLocation: r.CurLocation,
	}
	return robotLog
}

func (r *RobotStruct) LocateLog() (*os.File, error) {
	file, err := os.Open(r.Logname)
	return file, err
}
