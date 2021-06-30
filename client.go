package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

//GameData represents data received from server
type GameData struct {
	GameID   string
	Player1  string
	Player2  string
	Status   bool
	P1Field  [][]int
	P2Field  [][]int
	TurnFlag int
}

type Cell struct {
	Button *widget.Button
	X      int
	Y      int
}

type Ship struct {
	Size             int
	Orientation      string
	BaseDeckPosition [2]int //position of top/left corner of ship
}

type Fleet struct {
	TotalDecks int
	Size       map[string]int
	Array      []Ship
}

/*Requests to be sent:
GET 	http://sb.mailboxly.info/?gameid=AA22DD5511	-request to get data about game AA22DD5511
GET		http://sb.mailboxly.info/					-(???) request to get IDs of all current games
POST	http://sb.mailboxly.info/					-request to create new game room
PUT		http://sb.mailboxly.info/?gameid=AA22DD5511	-request to edit data in game AA22DD5511
*/

func main() {
	initGUI()
}

func initGUI() {

	window := newWindow()

	newMainContainer(window)

	window.ShowAndRun()
}

//Creates and sends new 'method' request to 'uri'. If used to create GET request dataToSend should
//be NIL, because GET request requires no data to send. Method returns not unmarshaled response body,
//so it means to use json.Unmarshal() next to sendRequest()
func sendRequest(method string, uri string, dataToSend []byte) []byte {
	client := &http.Client{}
	var err error
	var request *http.Request
	var response *http.Response

	request, err = http.NewRequest(method, uri, bytes.NewBuffer(dataToSend))
	if err != nil {
		fmt.Println(err)
	}

	//Separate method GET because it is the only method that requires no data to send, but only uri
	if method == "GET" {
		response, err = client.Get(uri)
	} else {
		response, err = client.Do(request)
	}

	if err != nil {
		fmt.Println(err)
	}

	json, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err)
	}

	return json
}

//Initializes new main container, which contains player's nickname (for yet) and 'Start game' button
//which sends all registration info to server to create a new game room
func newMainContainer(window fyne.Window) {
	var gameData GameData
	var fleet Fleet = Fleet{
		Size: make(map[string]int, 4),
	}

	usernameLabel := widget.NewLabel("Username: ")
	usernameEntry := widget.NewEntry()
	usernameRow := container.NewGridWithRows(2, usernameLabel, usernameEntry)

	userContainer := container.NewAdaptiveGrid(10)
	userContainer.Resize(fyne.NewSize(250, 250))

	shipsRadioGroup := widget.NewRadioGroup(
		[]string{"Single-deck ship", "Double-deck ship", "Three-deck ship", "Four-deck ship"},
		func(s string) {})
	shipsOrientation := widget.NewRadioGroup(
		[]string{"Horizontal", "Vertical"}, func(s string) {})
	shipsRadioGroup.SetSelected("Four-deck ship")
	shipsOrientation.SetSelected("Vertical")
	shipsContainer := container.NewVBox(shipsRadioGroup, widget.NewSeparator(), shipsOrientation)

	userCellArray := setButtons(userContainer, "putShip", [10][10]string{}, &fleet)

	//When the button is clicked, it sends POST request to create new game room
	startGameButton := widget.NewButton("Start game", func() {
		if fleet.TotalDecks == 20 {
			//response := sendRequest("GET", "http://localhost:8080/?gameid=AA22DD5511", nil)
			response := sendRequest("GET", "http://localhost:8080/game", nil)
			err := json.Unmarshal([]byte(response), &gameData)

			if err != nil {
				fmt.Println(err)
			} else {
				var buttonValuesArray [10][10]string
				for x := 0; x < 10; x++ {
					for y := 0; y < 10; y++ {
						buttonValuesArray[x][y] = userCellArray[x][y].Button.Text
					}
				}

				newGameContainer(window, gameData, buttonValuesArray)
			}
		} else {
			fmt.Println("Your ships set uncorrectly")
		}
	})
	startGameButton.Resize(fyne.NewSize(150, 50))

	mainContainer := container.NewWithoutLayout(usernameRow, userContainer, startGameButton, shipsContainer)
	mainContainer.Resize(fyne.NewSize(700, 500))

	verticalCenter := mainContainer.Size().Width / 2
	usernameRow.Move(fyne.NewPos(verticalCenter-usernameRow.Size().Width/2, 30))
	userContainer.Move(fyne.NewPos(verticalCenter-userContainer.Size().Width/2, 130))
	startGameButton.Move(fyne.NewPos(verticalCenter-startGameButton.Size().Width/2, mainContainer.Size().Height-100))
	shipsContainer.Move(fyne.NewPos(30, userContainer.Position().Y))

	window.SetContent(mainContainer)
	window.SetTitle("Sea Battle")
}

//Initializes new game container, which contains game details: both user and bot field,
//'End game' button (for yet), to close current game and open new main container
func newGameContainer(window fyne.Window, gameData GameData, buttonValuesArray [10][10]string) {
	//Size of cell fields depends on size of window
	fieldSize := window.Canvas().Size().Width/2 - 75

	userContainer := container.NewAdaptiveGrid(10)
	botContainer := container.NewAdaptiveGrid(10)

	userContainer.Move(fyne.NewPos(50, 80))
	botContainer.Move(fyne.NewPos(fieldSize+100, 80))
	userContainer.Resize(fyne.NewSize(fieldSize, fieldSize))
	botContainer.Resize(fyne.NewSize(fieldSize, fieldSize))

	player1Label := widget.NewLabel(gameData.Player1 + "'s field:")
	player2Label := widget.NewLabel(gameData.Player2 + "'s field:")
	player1Label.Move(fyne.NewPos(50, 30))
	player2Label.Move(fyne.NewPos(fieldSize+100, 30))

	//Setting cells in fields
	//userCellArray := setButtons(userContainer, "", buttonValuesArray)
	//botCellArray := setButtons(botContainer, "shoot", [10][10]string{})
	//fmt.Sprintf("%s", botCellArray, userCellArray)

	endGameButton := widget.NewButton("End game", func() {
		newMainContainer(window)
	})
	endGameButton.Move(fyne.NewPos(window.Canvas().Size().Width/2-50, window.Canvas().Size().Height-100))
	endGameButton.Resize(fyne.NewSize(100, 50))

	gameContainer := container.NewWithoutLayout(
		player1Label,
		player2Label,
		userContainer,
		botContainer,
		endGameButton,
	)

	//Adding containers to window
	window.SetContent(gameContainer)

	window.SetTitle("Sea Battle: Game ID: " + gameData.GameID)
}

//Sets a new window
//(Size should be set in both newWindow() and newMainContainer() methods)
func newWindow() fyne.Window {
	application := app.New()
	window := application.NewWindow("")
	window.Resize(fyne.NewSize(700, 500))
	window.SetFixedSize(true)
	window.CenterOnScreen()

	return window
}

//Sets cells on container. Parameter 'field' determines on which field cells should be set
//func setButtons(container *fyne.Container, handler func(), initTextValues [10][10]string) [10][10]Cell {
func setButtons(container *fyne.Container, listener string, initTextValues [10][10]string, fleet *Fleet) [10][10]Cell {
	var cellArray [10][10]Cell = [10][10]Cell{}

	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			cell := Cell{
				X: x,
				Y: y,
			}

			cell.Button = widget.NewButton(initTextValues[x][y], func() {
				switch listener {
				case "putShip":
					putShip(&cellArray, cell, fleet)
					container.Refresh()
				case "shoot":
					//PUT request
				}
			})

			/*//Marshalling cell's coordinates into dataToSend (type []byte)
			dataToSend, err := json.Marshal([]int{cell.X, cell.Y}) //may have problems
			if err != nil {
				fmt.Println(err)
			}
			rawData := sendRequest("PUT", "http://sb.mailboxly.info/?gameid=AA22DD5511", dataToSend)
			//Unmarshal received data from PUT request*/

			container.Add(cell.Button)
			cellArray[x][y] = cell
		}
	}

	return cellArray
}

//Puts a piece of ship in pressed cell if this cell is valid
func putShip(cellArray *[10][10]Cell, cell Cell, fleet *Fleet) {
	//delete ship if pressed cell is BaseDeck (left/top piece of ship)
	//if pressed cell have "O" text but it is not BaseDeck, call will be ignored
	if cell.Button.Text == "O" {
		eraseShip(cell, cellArray, fleet)
		return

		//check if cells around pressed cell are clear
		//if branch - cells aren't clear; terminate method
		//else branch - cells are clear; next validation stage
	} else if !cell.cellsAroundAreClear(cellArray) {
		return

		//check if supposed ship colliding something
		//if branch - ship collided something; terminate method
		//else branch - cell around supposed ship are clear; next validation stage
	} else if cell.shipCollided(cellArray, "Three-deck ship", "Horizontal") {
		return

	} else {
		drawShip(newShip("Horizontal", "Three-deck ship", cell, fleet), cellArray)
		fmt.Println("\n", fleet.Size)
		fmt.Println(fleet.Array)
		fmt.Println(fleet.TotalDecks)
	}
}

//Deletes ship from both field and Fleet struct
func eraseShip(cell Cell, cellArray *[10][10]Cell, fleet *Fleet) {
	x := cell.X
	y := cell.Y

	for _, ship := range fleet.Array {
		if ship.BaseDeckPosition[0] == x && ship.BaseDeckPosition[1] == y {
			switch ship.Size {
			case 1:
				fleet.Size["Single-deck ship"] -= 1
			case 2:
				fleet.Size["Double-deck ship"] -= 1
			case 3:
				fleet.Size["Three-deck ship"] -= 1
			case 4:
				fleet.Size["Four-deck ship"] -= 1
			}

			fleet.TotalDecks -= ship.Size

			for i := 0; i < ship.Size; i++ {
				switch ship.Orientation {
				case "Vertical":
					cellArray[x+i][y].Button.Text = ""
				case "Horizontal":
					cellArray[x][y+i].Button.Text = ""
				}
			}

			//Removing ship from slice
			shipArrayCopy := fleet.Array
			fleet.Array = nil
			for _, s := range shipArrayCopy {
				if s != ship {
					fleet.Array = append(fleet.Array, s)
				}
			}

			break
		}
	}
}

//Draws a ship in cell array
func drawShip(ship Ship, cellArray *[10][10]Cell) {
	x := ship.BaseDeckPosition[0]
	y := ship.BaseDeckPosition[1]

	for i := 0; i < ship.Size; i++ {
		switch ship.Orientation {
		case "Horizontal":
			cellArray[x][y+i].Button.Text = "O"
		case "Vertical":
			cellArray[x+i][y].Button.Text = "O"
		}
	}
}

//Creates new ship with given parameter
func newShip(orientation string, size string, cell Cell, fleet *Fleet) Ship {
	var ship Ship

	switch size {
	case "Single-deck ship":
		ship.Size = 1
	case "Double-deck ship":
		ship.Size = 2
	case "Three-deck ship":
		ship.Size = 3
	case "Four-deck ship":
		ship.Size = 4
	}

	ship.BaseDeckPosition = [2]int{cell.X, cell.Y}
	ship.Orientation = orientation

	fleet.Size[size] += 1
	fleet.TotalDecks += ship.Size
	fleet.Array = append(fleet.Array, ship)

	return ship
}

//Validation method. Returns true if cells around hit cell are clear, and false if cells aren't
func (cell Cell) cellsAroundAreClear(cellArray *[10][10]Cell) bool {
	for x := -1; x <= 1; x++ {
		for y := -1; y <= 1; y++ {
			//going to next iteration if cell.X is 0 or 9
			//to get rid of index out of range exception
			if cell.X == 0 && x == -1 || cell.X == 9 && x == 1 {
				break
			}

			//going to next iteration if cell.Y is 0 or 9
			//to get rid of index out of range exception
			if cell.Y == 0 && y == -1 || cell.Y == 9 && y == 1 {
				continue
			}

			//false returns if atleast one of corners around cell have "O" text
			if cellArray[cell.X+x][cell.Y+y].Button.Text == "O" {
				return false
			}
		}
	}

	return true
}

//Validation method. Returns true if ship collided something, and false if ship not
func (cell Cell) shipCollided(cellArray *[10][10]Cell, shipSize string, shipOrientation string) bool {
	var size int

	switch shipSize {
	case "Single-deck ship":
		size = 1
	case "Double-deck ship":
		size = 2
	case "Three-deck ship":
		size = 3
	case "Four-deck ship":
		size = 4
	}

	var validationResult bool
	for i := 1; i < size; i++ {
		if shipOrientation == "Vertical" {
			validationResult = cellArray[cell.X+i][cell.Y].cellsAroundAreClear(cellArray)
		} else if shipOrientation == "Horizontal" {
			validationResult = cellArray[cell.X][cell.Y+i].cellsAroundAreClear(cellArray)
		}
	}

	return false || !validationResult
}
