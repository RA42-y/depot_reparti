package site

import (
	"fmt"
	"github.com/Szzx123/depot_reparti/service"
	"github.com/gorilla/websocket"
	"log"
	"os"
	"strconv"
	"sync"

	"github.com/Szzx123/depot_reparti/global"
	"github.com/Szzx123/depot_reparti/model/message"
	"github.com/Szzx123/depot_reparti/utils"
)

var (
	mutex = &sync.Mutex{}
)

type Site struct {
	Num string
}

var (
	websocketConn *websocket.Conn
)

func New_Site(num string) *Site {
	return &Site{
		Num: num,
	}
}

func (site *Site) Run() {
	go site.Message_Interceptor()
}

func (site *Site) Message_Interceptor() {
	// var to_operate_cargo bool = false
	l := log.New(os.Stderr, "", 0)
	l.Println("Start of message_interceptor ", site.Num)
	for {
		var rcv_msg, cargo, msg_type, operation, receiver, sender, horloge_snapshot, snapshot string
		var quantity, stock_A, stock_B, stock_C int
		fmt.Scanln(&rcv_msg)
		mutex.Lock() // treat champ sender
		sender = utils.Findval(rcv_msg, "sender")
		if sender != "C"+site.Num[1:] {
			mutex.Unlock()
			continue
		}

		// treat champ receiver
		receiver = utils.Findval(rcv_msg, "receiver")
		if receiver != site.Num {
			mutex.Unlock()
			continue
		}
		l.Printf("site %s received message: %s", site.Num, rcv_msg)
		// treat champ type
		msg_type = utils.Findval(rcv_msg, "type")
		switch msg_type {
		case "débutSC":
		case "updateSC":
		case "generateSnapshot":
		default:
			l.Println("Unknown message type")
			mutex.Unlock()
			continue
		}

		// treat champ cargo
		cargo = utils.Findval(rcv_msg, "cargo")

		// treat champ operation
		operation = utils.Findval(rcv_msg, "operation")

		// treat champ quantity
		quantity_string := utils.Findval(rcv_msg, "quantity")
		if quantity_string != "" {
			quantity, _ = strconv.Atoi(quantity_string)
		}

		stock_A_string := utils.Findval(rcv_msg, "A")
		if stock_A_string != "" {
			stock_A, _ = strconv.Atoi(stock_A_string)
		}

		stock_B_string := utils.Findval(rcv_msg, "B")
		if stock_B_string != "" {
			stock_B, _ = strconv.Atoi(stock_B_string)
		}

		stock_C_string := utils.Findval(rcv_msg, "C")
		if stock_C_string != "" {
			stock_C, _ = strconv.Atoi(stock_C_string)
		}

		horloge_snapshot = utils.Findval(rcv_msg, "horloge_snapshot")

		snapshot = utils.Findval(rcv_msg, "snapshot")

		msg_to_handle := message.New_SiteMessage(msg_type, cargo, operation, quantity, stock_A, stock_B, stock_C, horloge_snapshot, snapshot)
		site.Message_Handler(*msg_to_handle)
		mutex.Unlock()
	}
}

func (site *Site) Message_Handler(msg message.SiteMessage) {
	l := log.New(os.Stderr, "", 0) // test
	switch msg.TypeMessage {
	case "débutSC":
		// 操作库存
		if msg.Operation == "in" {
			global.Depot.Cargo_IN(msg.Cargo, msg.Quantity)
		} else if msg.Operation == "out" {
			global.Depot.Cargo_OUT(msg.Cargo, msg.Quantity)
		}
		// l.Println(global.Depot) // test
		// 发送finSC给ctl
		utils.Msg_send(utils.Msg_format("receiver", "C"+site.Num[1:]) + utils.Msg_format("type", "finSC") + utils.Msg_format("sender", site.Num) + utils.Msg_format("A", strconv.Itoa(global.Depot.StoreHouse["A"])) + utils.Msg_format("B", strconv.Itoa(global.Depot.StoreHouse["B"])) + utils.Msg_format("C", strconv.Itoa(global.Depot.StoreHouse["C"])))
		msgCargo := message.CargoMessage{Site: site.Num, TypeMessage: "updateCargo", Stock_A: global.Depot.Get_Cargo("A"), Stock_B: global.Depot.Get_Cargo("B"), Stock_C: global.Depot.Get_Cargo("C")}
		service.Cargo_Send_Handler(msgCargo)
	case "updateSC":
		// 更新所有库存信息
		global.Depot.Set_Cargo("A", msg.Stock_A)
		global.Depot.Set_Cargo("B", msg.Stock_B)
		global.Depot.Set_Cargo("C", msg.Stock_C)
		l.Println(site.Num, ":", global.Depot)
		msgCargo := message.CargoMessage{Site: site.Num, TypeMessage: "updateCargo", Stock_A: msg.Stock_A, Stock_B: msg.Stock_B, Stock_C: msg.Stock_C}
		service.Cargo_Send_Handler(msgCargo)
	case "generateSnapshot":
		msgSnapshot := message.SnapshotMessage{Site: site.Num, TypeMessage: "generateSnapshot", Horloge: msg.HorlogeSnapshot, Snapshot: msg.Snapshot}
		service.Snapshot_Send_Handler(msgSnapshot)
	}
}
