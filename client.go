package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

//GameData represents data received from server
type GameData struct {
	GameID     string
	Player1    string
	Player2    string
	ShotStatus string
	UserX      int
	UserY      int
	BotX       int
	BotY       int
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

var gameData GameData
var fleet Fleet = Fleet{
	Size: make(map[string]int, 4),
}

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
func sendRequest(method string, uri string, rawDataToSend interface{}) []byte {
	client := &http.Client{}
	var err error
	var request *http.Request
	var response *http.Response
	var dataToSend []byte

	if dataToSend, err = json.Marshal(rawDataToSend); err != nil {
		fmt.Println(err)
	}

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
	var mainContainer *fyne.Container

	usernameLabel := widget.NewLabel("Username: ")
	usernameEntry := widget.NewEntry()
	usernameRow := container.NewGridWithRows(2, usernameLabel, usernameEntry)

	userContainer := container.NewAdaptiveGrid(10)
	userContainer.Resize(fyne.NewSize(250, 250))

	shipsSize := widget.NewRadioGroup(
		[]string{"Single-deck ship", "Double-deck ship", "Three-deck ship", "Four-deck ship"},
		func(s string) {})
	shipsOrientation := widget.NewRadioGroup(
		[]string{"Horizontal", "Vertical"},
		func(s string) {})
	shipsSize.SetSelected("Four-deck ship")
	shipsOrientation.SetSelected("Vertical")
	shipsContainer := container.NewVBox(shipsSize, widget.NewSeparator(), shipsOrientation)

	userCellArray := setButtons(userContainer, "putShip", &fleet, shipsOrientation, shipsSize)

	//When the button is clicked, it sends POST request to create new game room
	startGameButton := widget.NewButton("Start game", func() {
		if fleet.TotalDecks == 20 && usernameEntry.Text != "" {
			response := sendRequest("POST", "http://localhost:8080/", usernameEntry.Text)

			if err := json.Unmarshal(response, &gameData); err != nil {
				fmt.Println(err)
			} else {
				newGameContainer(window, gameData, fleet)
			}
		} else if fleet.TotalDecks < 20 {
			fmt.Println("\nYour fleet is not complete")
		} else {
			fmt.Println("\nYou haven't entered your nickname")
		}
	})
	startGameButton.Resize(fyne.NewSize(150, 50))

	randomShipButton := widget.NewButton("Random ships", func() {
		setFleetAutomatically(&userCellArray, &fleet)
		mainContainer.Refresh()
	})
	randomShipButton.Resize(fyne.NewSize(150, 50))

	mainContainer = container.NewWithoutLayout(usernameRow, userContainer, startGameButton, randomShipButton, shipsContainer)
	mainContainer.Resize(fyne.NewSize(700, 500))

	verticalCenter := mainContainer.Size().Width / 2
	usernameRow.Move(fyne.NewPos(verticalCenter-usernameRow.Size().Width/2, 30))
	userContainer.Move(fyne.NewPos(verticalCenter-userContainer.Size().Width/2, 130))
	startGameButton.Move(fyne.NewPos(verticalCenter-startGameButton.Size().Width/2, mainContainer.Size().Height-100))
	shipsContainer.Move(fyne.NewPos(30, userContainer.Position().Y))
	randomShipButton.Move(fyne.NewPos(500, userContainer.Position().Y))

	window.SetContent(mainContainer)
	window.SetTitle("Sea Battle")
}

//Initializes new game container, which contains game details: both user and bot field,
//'End game' button (for yet), to close current game and open new main container
func newGameContainer(window fyne.Window, gameData GameData, fleet Fleet) {
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
	userCellArray := setButtons(userContainer, "", nil, nil, nil)
	botCellArray := setButtons(botContainer, "shoot", nil, nil, nil)
	fmt.Sprintf("%s", botCellArray, userCellArray)

	for _, ship := range fleet.Array {
		drawShip(ship, &userCellArray)
	}

	endGameButton := widget.NewButton("End game", func() {
		sendRequest("DELETE", "http://localhost:8080/game", gameData.GameID)
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
	gameContainer.Refresh()
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
func setButtons(container *fyne.Container, listener string, fleet *Fleet,
	shipOrientation *widget.RadioGroup, shipSize *widget.RadioGroup) [10][10]Cell {

	var cellArray [10][10]Cell = [10][10]Cell{}

	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			cell := Cell{
				X: x,
				Y: y,
			}

			cell.Button = widget.NewButton("", func() {
				switch listener {
				case "putShip":
					validateAreaForShip(&cellArray, cell, fleet,
						shipOrientation.Selected, shipSize.Selected)

					container.Refresh()
				case "shoot":
					if cell.Button.Text == "" {
						shoot(cell, cellArray)
						analyzeResponse(&cellArray, &cell)
	
						container.Refresh()
					} else {
						fmt.Println("\nYou were shooting this cell already")
					}
				}
			})

			container.Add(cell.Button)
			cellArray[x][y] = cell
		}
	}

	return cellArray
}

//Analyzes response from server and draws true data on bot's field
func analyzeResponse(cellArray *[10][10]Cell, cell *Cell) {
	switch gameData.ShotStatus {
	case "miss":
		cell.Button.Text = "*"
	case "hit":
		cell.Button.Text = "X"
	case "kill":
		cell.Button.Text = "X"
		coverKilledShip(cellArray, cell)
	}
}

func coverKilledShip(cellArray *[10][10]Cell, cell *Cell) {
	shipOrientation := getKilledShipOrientation(cellArray, cell)
	decksBehind, decksInFront := getKilledShipSize(cellArray, cell, shipOrientation)

	fmt.Println("Orientation:", shipOrientation, "InFront:", decksInFront, "Behind:", decksBehind)

	coverOneDeck(cellArray, cell.X, cell.Y)

	for step := 1; step <= decksBehind; step++ {
		switch shipOrientation {
		case "Vertical":
			coverOneDeck(cellArray, cell.X-step, cell.Y)
		case "Horizontal":
			coverOneDeck(cellArray, cell.X, cell.Y-step)
		}
	}

	for step := 1; step <= decksInFront; step++ {
		switch shipOrientation {
		case "Vertical":
			coverOneDeck(cellArray, cell.X+step, cell.Y)
		case "Horizontal":
			coverOneDeck(cellArray, cell.X, cell.Y+step)
		}
	}

}

//Returns orientation of killed ship
func getKilledShipOrientation(cellArray *[10][10]Cell, cell *Cell) string {
	if cell.cellsAroundAreClear(cellArray) {
		return "Single-deck ship"
	}

	if cell.X + 1 <= 9 {
		if cellArray[cell.X+1][cell.Y].Button.Text == "X" {
			return "Vertical"
		}
	}
	if cell.X - 1 >= 0 {
		if cellArray[cell.X-1][cell.Y].Button.Text == "X" {
			return "Vertical"
		}
	}
	if cell.Y + 1 <= 9 {
		if cellArray[cell.X][cell.Y+1].Button.Text == "X" {
			return "Horizontal"
		}
	} 
	if cell.Y - 1 >= 0 {
		if cellArray[cell.X][cell.Y-1].Button.Text == "X" {
			return "Horizontal"
		}
	}
	
	return ""
}

//Returns two values: number of cell that located behind last hit cell of killed ship
//and number of cells that located in front of last hit cell of killed ship. Last hit cell
//not included in both values
func getKilledShipSize(cellArray *[10][10]Cell, cell *Cell, shipOrientation string) (int, int) {
	var decksInFront int	//number of decks that located in front of last hit cell of killed ship
	var decksBehind int 	//number of decks that located behind last hit cell of killed ship

	for step := 1; step <= 4; step++{
		if step == 4 {
			break
		}

		switch shipOrientation {
			case "Vertical": {
				if cell.X + step <= 9 {
					if cellArray[cell.X + step][cell.Y].Button.Text == "X" {
						decksInFront++
					}
				}
				if cell.X - step >= 0 {
					if cellArray[cell.X - step][cell.Y].Button.Text == "X" {
						decksBehind++
					}
				}
			}
			
			case "Horizontal": {
				if cell.Y + step <= 9 {
					if cellArray[cell.X][cell.Y + step].Button.Text == "X" {
						decksInFront++
					}
				}
				if cell.Y - step >= 0 {
					if cellArray[cell.X][cell.Y - step].Button.Text == "X" {
						decksBehind++
					}
				}
			}
		}
	}
	return decksBehind, decksInFront
}

//Covers one piece of killed ship
func coverOneDeck(cellArray *[10][10]Cell, cellX int, cellY int) {
	for x := -1; x <= 1; x++ {
		for y := -1; y <= 1; y++ {
			if cellX == 0 && x == -1 || cellX == 9 && x == 1 {
				break
			}

			if cellY == 0 && y == -1 || cellY == 9 && y == 1 {
				continue
			}

			if x == 0 && y == 0 {
				cellArray[cellX+x][cellY+y].Button.Text = "X"
			} else if cellArray[cellX+x][cellY+y].Button.Text != "X" {
				cellArray[cellX+x][cellY+y].Button.Text = "*"
			}
		}
	}
}

//Sends PUT request to server when user hits cell in bot's field.
//Response saved in GameData struct
func shoot(cell Cell, cellArray [10][10]Cell) {
	gameData.UserX = cell.X
	gameData.UserY = cell.Y
	response := sendRequest("PUT", "http://localhost:8080/game", gameData)

	//Unmarshal received data from PUT request
	if err := json.Unmarshal(response, &gameData); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(gameData)
	}
}

//Handler for 'Random ships' button. Clears all current data about user fleet
//and sets new ships with random location and orientation
func setFleetAutomatically(cellArray *[10][10]Cell, fleet *Fleet) {
	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			eraseShip(cellArray[x][y], cellArray, fleet)
		}
	}

	var shipOrientation string
	var shipSize string

	rand.Seed(int64(time.Now().Nanosecond()))

	for fleet.TotalDecks != 20 && len(fleet.Array) < 10 {
		switch rand.Intn(2) {
		case 0:
			shipOrientation = "Vertical"
		case 1:
			shipOrientation = "Horizontal"
		}

		shipSize = whichShipToSet(fleet)

		x := rand.Intn(10)
		y := rand.Intn(10)

		validateAreaForShip(cellArray, cellArray[x][y], fleet, shipOrientation, shipSize)
	}
}

//Analyzes current fleet and returns which ship should be set now
func whichShipToSet(fleet *Fleet) string {
	var singleDeck, doubleDeck, threeDeck, fourDeck int

	for _, ship := range fleet.Array {
		switch ship.Size {
		case 1:
			singleDeck++
		case 2:
			doubleDeck++
		case 3:
			threeDeck++
		case 4:
			fourDeck++
		}
	}

	if fourDeck < 1 {
		return "Four-deck ship"
	} else if threeDeck < 2 {
		return "Three-deck ship"
	} else if doubleDeck < 3 {
		return "Double-deck ship"
	} else if singleDeck < 4 {
		return "Single-deck ship"
	}

	return ""
}

//Validates area for supposed ship. Each stage has it's own job: to check if ship
//collides another ship or field borders; or to check if fleet have empty space for
//supposed ship. Also method can delete existing ship if user hits cell, which is base
//deck of existing ship.
//func validateAreaForShip(cellArray *[10][10]Cell, cell Cell, fleet *Fleet) {
func validateAreaForShip(cellArray *[10][10]Cell, cell Cell,
	fleet *Fleet, shipOrientation string, shipSize string) {

	//terminate method if something gone wrong with ship's parameters
	if shipOrientation == "" || shipSize == "" {
		fmt.Println("\nSize and/or orientation values are empty")
		return
	}

	//delete ship if pressed cell is BaseDeck (left/top piece of ship)
	//if pressed cell have "#" text but it is not BaseDeck, call will be ignored
	if cell.Button.Text == "^" || cell.Button.Text == "<" {
		eraseShip(cell, cellArray, fleet)
		return

		//check if supposed ship colliding something
		//if branch - ship collided something; terminate method
		//else branch - cell around supposed ship are clear; next validation stage
		//Backup params: cellArray, "Single-deck ship", "Horizontal"
	} else if !cell.shipCollision(cellArray, shipSize, shipOrientation) {
		return

		//check if fleet have free space for supposed ship
		//if branch - there aren't free space; terminate method
		//else branch - there free space; ship can be placed
		//Backup params: "Single-deck ship", *fleet
	} else if !fleetHaveFreeSpace(shipSize, *fleet) {
		return

		//program draws specified ship on field if all validations above allows to
		//Backup params: "Horizontal", "Single-deck ship", cell, fleet), cellArray
	} else {
		drawShip(newShip(shipOrientation, shipSize, cell, fleet), cellArray)
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
			{
				if i == 0 {
					cellArray[x][y+i].Button.Text = "<"
				} else {
					cellArray[x][y+i].Button.Text = "#"
				}
			}
		case "Vertical":
			{
				if i == 0 {
					cellArray[x+i][y].Button.Text = "^"
				} else {
					cellArray[x+i][y].Button.Text = "#"
				}
			}
		}
	}
}

//Creates new ship with given parameters
func newShip(shipOrientation string, shipSize string, cell Cell, fleet *Fleet) Ship {
	var ship Ship

	switch shipSize {
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
	ship.Orientation = shipOrientation

	fleet.Size[shipSize] += 1
	fleet.TotalDecks += ship.Size
	fleet.Array = append(fleet.Array, ship)

	return ship
}

//Validation method. Returns true if cells around hit cell are clear, and false if cells aren't
func (cell Cell) cellsAroundAreClear(cellArray *[10][10]Cell) bool {
	for x := -1; x <= 1; x++ {
		for y := -1; y <= 1; y++ {
			if x == 0 && y == 0 {
				continue
			}

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

			//false returns if atleast one of cells around hit cell have:
			//"#" or "<" or "^" text when we are working on user's field
			//"X" when we are working on bot's field
			text := cellArray[cell.X+x][cell.Y+y].Button.Text
			if text == "#" || text == "<" || text == "^" || text == "X" {
				return false
			}
		}
	}

	return true
}

//Validation method. Returns false if ship collided something,
//and true if ship doesn't and can be placed in area
func (cell Cell) shipCollision(cellArray *[10][10]Cell, shipSize string, shipOrientation string) bool {
	var size int
	var validationResult bool

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

	for i := 0; i < size; i++ {
		if shipOrientation == "Vertical" {
			if cell.X+i <= 9 {
				validationResult = cellArray[cell.X+i][cell.Y].cellsAroundAreClear(cellArray)
			} else {
				validationResult = false
			}
		} else if shipOrientation == "Horizontal" {
			if cell.Y+i <= 9 {
				validationResult = cellArray[cell.X][cell.Y+i].cellsAroundAreClear(cellArray)
			} else {
				validationResult = false
			}
		}

		if validationResult == false {
			break
		}
	}

	return validationResult
}

//Validation method. Returns true if number of existing ships
//with same size is not greater than maximum amount.
//Return false if there are no free space in fleet's
//'Size map[string]int' or if error occured
func fleetHaveFreeSpace(shipSize string, fleet Fleet) bool {
	var maxAmount int //used to contain maximum number for specified type of ship

	switch shipSize {
	case "Single-deck ship":
		maxAmount = 4
	case "Double-deck ship":
		maxAmount = 3
	case "Three-deck ship":
		maxAmount = 2
	case "Four-deck ship":
		maxAmount = 1
	default: //default used if something go wrong with value of ship's size
		return false
	}

	if fleet.Size[shipSize] < maxAmount {
		return true
	}

	return false
}
