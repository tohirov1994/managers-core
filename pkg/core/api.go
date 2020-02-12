package core

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	DSN "github.com/tohirov1994/database"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type managersStruct struct {
	Id       int
	Name     string
	Surname  string
	Login    string
	Password string
}

type clientsStruct struct {
	Id       int
	Name     string
	Surname  string
	Login    string
	Password string
}

type clientsCardsStruct struct {
	Id         int
	PAN        int
	PIN        int
	Balance    int
	HolderName string
	CVV        int
	Validity   int
	ClientId   int
}

type ATMStruct struct {
	Id       int
	City     string
	District string
	Street   string
}

type servicesStruct struct {
	Id      int
	Service string
	Balance int
}

var PassWrong = errors.New("password is not valid")

func Init(db *sql.DB) (err error) {
	initDDLsDMLs := []string{DSN.ManagersDDL, DSN.ClientsDDL, DSN.ClientsCardsDDL, DSN.AtmsDDL, DSN.ServicesDDL,
		DSN.ManagersDML, DSN.ClientsDML, DSN.ClientsCardsDML, DSN.AtmsDML, DSN.ServicesDML}
	for _, init := range initDDLsDMLs {
		_, err = db.Exec(init)
		if err != nil {
			return err
		}
	}
	return nil
}

func ATMsGet(db *sql.DB) (ATMs []ATMStruct, err error) {
	rows, err := db.Query(DSN.GetATMData)
	if err != nil {
		return nil, err
	}
	defer func() {
		if innerErr := rows.Close(); innerErr != nil {
			ATMs = nil
		}
	}()
	for rows.Next() {
		atm := ATMStruct{}
		err = rows.Scan(&atm.Id, &atm.City, &atm.District, &atm.Street)
		if err != nil {
			return nil, err
		}
		ATMs = append(ATMs, atm)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return ATMs, nil
}

func SignIn(loginUsr, passwordUsr string, db *sql.DB) (bool, error) {
	var dbLogin, dbPassword string
	err := db.QueryRow(DSN.GetLoginPassManager, loginUsr).Scan(&dbLogin, &dbPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	if dbPassword != passwordUsr {
		return false, PassWrong
	}
	return true, nil
}

func AddClient(nameClient, surnameClient, loginClient, passwordClient string, db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	_, err = tx.Exec(
		DSN.InsertClient,
		sql.Named("name", nameClient),
		sql.Named("surname", surnameClient),
		sql.Named("login", loginClient),
		sql.Named("password", passwordClient),
	)
	if err != nil {
		return err
	}
	return nil
}

func PANLastPlusOne(db *sql.DB) (pan int64, err error) {
	var lastPAN int64
	err = db.QueryRow(DSN.GetLastPAN, pan).Scan(&lastPAN)
	if err != nil {
		fmt.Printf("cant find last PAN Number %v", err)
		return 0, err
	}
	lastPAN = lastPAN + 1
	return lastPAN, nil
}

func CheckIdClient(checkId int64, db *sql.DB) (idAccept int64, err error) {
	db.QueryRow(DSN.CheckIdClient, checkId).Scan(&idAccept)
	if err != nil {
		fmt.Printf("can't find Client Id: %v", err)
		return 0, err
	}
	return idAccept, nil
}

func GetNameSurnameFromIdClient(idClient int64, db *sql.DB) (nameClient, surnameClient string, err error) {
	err = db.QueryRow(DSN.GetLoginPassManager, idClient).Scan(&nameClient, &surnameClient)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", err
		}
		return "", "", err
	}
	return nameClient, surnameClient, nil
}

func AddCardToClient(panCard, pinCard, balanceCard int64, holderNameCard string, cvvCard, validityCard, clientIdCard int64, db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}()
	_, err = tx.Exec(
		DSN.InsertClientCard,
		sql.Named("pan", panCard),
		sql.Named("pin", pinCard),
		sql.Named("balance", balanceCard),
		sql.Named("holderName", holderNameCard),
		sql.Named("cvv", cvvCard),
		sql.Named("validity", validityCard),
		sql.Named("clientId", clientIdCard),
	)
	if err != nil {
		return err
	}
	return nil
}

func AddServiceToTheBank(servicedName string, db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}()
	_, err = tx.Exec(
		DSN.InsertService,
		sql.Named("serviceName", servicedName),
		sql.Named("serviceBalance", 0),
	)
	if err != nil {
		return err
	}
	return nil
}

func AddAtmToTheBank(city, district, street string, db *sql.DB) (err error) {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
			return
		}
		err = tx.Commit()
	}()
	_, err = tx.Exec(
		DSN.InsertAtm,
		sql.Named("cityName", city),
		sql.Named("districtName", district),
		sql.Named("streetName", street),
	)
	if err != nil {
		return err
	}
	return nil
}

//TODO ############################### MARSHALING START HERE ###############################

//TODO ############################### CONVERTING SQL-DATA TO data-STRUCTURES ###############################
func DbManagersToStruct(db *sql.DB) (managers []managersStruct, err error) {
	rows, err := db.Query(DSN.GetManagerData)
	if err != nil {
		return nil, err
	}
	defer func() {
		if innerErr := rows.Close(); innerErr != nil {
			managers = nil
		}
	}()
	for rows.Next() {
		manager := managersStruct{}
		err = rows.Scan(&manager.Id, &manager.Name, &manager.Surname, &manager.Login, &manager.Password)
		if err != nil {
			return nil, err
		}
		managers = append(managers, manager)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return managers, nil
}

func DbClientsToStruct(db *sql.DB) (clients []clientsStruct, err error) {
	rows, err := db.Query(DSN.GetClientData)
	if err != nil {
		return nil, err
	}
	defer func() {
		if innerErr := rows.Close(); innerErr != nil {
			clients = nil
		}
	}()
	for rows.Next() {
		client := clientsStruct{}
		err = rows.Scan(&client.Id, &client.Name, &client.Surname, &client.Login, &client.Password)
		if err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return clients, nil
}

func DbClientsCardsToStruct(db *sql.DB) (clientsCards []clientsCardsStruct, err error) {
	rows, err := db.Query(DSN.GetCardsData)
	if err != nil {
		return nil, err
	}
	defer func() {
		if innerErr := rows.Close(); innerErr != nil {
			clientsCards = nil
		}
	}()
	for rows.Next() {
		clientCard := clientsCardsStruct{}
		err = rows.Scan(&clientCard.Id, &clientCard.PAN, &clientCard.PIN, &clientCard.Balance, &clientCard.HolderName, &clientCard.CVV, &clientCard.Validity, &clientCard.ClientId)
		if err != nil {
			return nil, err
		}
		clientsCards = append(clientsCards, clientCard)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return clientsCards, nil
}

func DbATMsToStruct(db *sql.DB) (ATMs []ATMStruct, err error) {
	rows, err := db.Query(DSN.GetATMData)
	if err != nil {
		return nil, err
	}
	defer func() {
		if innerErr := rows.Close(); innerErr != nil {
			ATMs = nil
		}
	}()
	for rows.Next() {
		atm := ATMStruct{}
		err = rows.Scan(&atm.Id, &atm.City, &atm.District, &atm.Street)
		if err != nil {
			return nil, err
		}
		ATMs = append(ATMs, atm)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return ATMs, nil
}

func DbServicesToStruct(db *sql.DB) (services []servicesStruct, err error) {
	rows, err := db.Query(DSN.GetServicesData)
	if err != nil {
		return nil, err
	}
	defer func() {
		if innerErr := rows.Close(); innerErr != nil {
			services = nil
		}
	}()
	for rows.Next() {
		service := servicesStruct{}
		err = rows.Scan(&service.Id, &service.Service, &service.Balance)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return services, nil
}

//TODO ############################### CONVERTING Data-STRUCTURES TO data[]byte ###############################
func ManagersDataStructToBytes(manager []managersStruct) (dataBytes []byte, err error) {
	log.Print("conversion data is started")
	dataStringMarshalIndent, err := json.MarshalIndent(manager, "", "   ")
	if err != nil {
		log.Fatalf("error to converted ManagersData to data[]Byte: %v", err)
	}
	return dataStringMarshalIndent, err
}

func ClientDataStructToBytes(client []clientsStruct) (dataBytes []byte, err error) {
	log.Print("conversion data is started")
	dataStringMarshalIndent, err := json.MarshalIndent(client, "", "   ")
	if err != nil {
		log.Fatalf("error to converted ClientData to data[]Byte: %v", err)
	}
	return dataStringMarshalIndent, err
}

func ClientsCardsDataStructToBytes(clientCard []clientsCardsStruct) (dataBytes []byte, err error) {
	log.Print("conversion data is started")
	dataStringMarshalIndent, err := json.MarshalIndent(clientCard, "", "   ")
	if err != nil {
		log.Fatalf("error to converted ClientsCardsData to data[]Byte: %v", err)
	}
	return dataStringMarshalIndent, err
}

func ATMsDataStructToBytes(ATM []ATMStruct) (dataBytes []byte, err error) {
	log.Print("conversion data is started")
	dataStringMarshalIndent, err := json.MarshalIndent(ATM, "", "   ")
	if err != nil {
		log.Fatalf("error to converted ATMsData to data[]Byte: %v", err)
	}
	return dataStringMarshalIndent, err
}

func ServicesDataStructToBytes(service []servicesStruct) (dataBytes []byte, err error) {
	log.Print("conversion data is started")
	dataStringMarshalIndent, err := json.MarshalIndent(service, "", "   ")
	if err != nil {
		log.Fatalf("error to converted ServicesData to data[]Byte: %v", err)
	}
	return dataStringMarshalIndent, err
}

//TODO ############################### WRITING DATA[]BYTE TO FILE PATH ###############################
func WriteToFileManagers(data []byte) (Result string, err error) {
	_ = os.Mkdir("backup", 0666)
	path := "backup/managers.json"
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		log.Print("file \"", path, "\" not exist for write")
		log.Print("file \"", path, "\" will be create")
		var file, err = os.Create(path)
		if err != nil {
			return "file \"managers.json\" could not be created successfully", err
		}
		defer file.Close()
		log.Print("file \"", path, "\" created successfully")
		log.Print("Data will saving")
		err = ioutil.WriteFile(path, data, 0666)
		if err != nil {
			log.Fatal(fmt.Errorf("can't save to %s: %w", path, err))
		}
		log.Print("Data was saved")
		return "Success", nil
	}
	log.Print("file ", path, " is exist")
	srcFile, err := os.Open(path)
	if err != nil {
		log.Fatal(fmt.Errorf("can't check source file %s: %w", path, err))
	}
	log.Print("source file was checked successfully")
	dateTimeBackup := time.Now().Format("backup/managersDataBackup(01-02-2006-15-04-5).json")
	log.Print("copy date and time for copy")
	fmt.Println(dateTimeBackup)
	log.Print("try create file to backup")
	destFile, err := os.Create(dateTimeBackup)
	if err != nil {
		log.Fatal(fmt.Errorf("can't create file to backup %s: %w", path, err))
	}
	log.Print("backup file was create")
	defer destFile.Close()
	log.Print("try copy from file to backup file")
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		log.Fatal(fmt.Errorf("can't finish copying %s: %w", path, err))
	}
	log.Print("file was copy to backup file")
	log.Print("Data will saving")
	err = ioutil.WriteFile(path, data, 0666)
	if err != nil {
		log.Fatal(fmt.Errorf("can't save to %s: %w", path, err))
	}
	log.Print("Data was saved")
	return "Success", nil
}

func WriteToFileClients(data []byte) (Result string, err error) {
	_ = os.Mkdir("backup", 0666)
	path := "backup/clients.json"
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		log.Print("file \"", path, "\" not exist for write")
		log.Print("file \"", path, "\" will be create")
		var file, err = os.Create(path)
		if err != nil {
			return "file \"clients.json\" could not be created successfully", err
		}
		defer file.Close()
		log.Print("file \"", path, "\" created successfully")
		log.Print("Data will saving")
		err = ioutil.WriteFile(path, data, 0666)
		if err != nil {
			log.Fatal(fmt.Errorf("can't save to %s: %w", path, err))
		}
		log.Print("Data was saved")
		return "Success", nil
	}
	log.Print("file ", path, " is exist")
	srcFile, err := os.Open(path)
	if err != nil {
		log.Fatal(fmt.Errorf("can't check source file %s: %w", path, err))
	}
	log.Print("source file was checked successfully")
	dateTimeBackup := time.Now().Format("backup/clientsDataBackup(01-02-2006-15-04-5).json")
	log.Print("copy date and time for copy")
	fmt.Println(dateTimeBackup)
	log.Print("try create file to backup")
	destFile, err := os.Create(dateTimeBackup)
	if err != nil {
		log.Fatal(fmt.Errorf("can't create file to backup %s: %w", path, err))
	}
	log.Print("backup file was create")
	defer destFile.Close()
	log.Print("try copy from file to backup file")
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		log.Fatal(fmt.Errorf("can't finish copying %s: %w", path, err))
	}
	log.Print("file was copy to backup file")
	log.Print("Data will saving")
	err = ioutil.WriteFile(path, data, 0666)
	if err != nil {
		log.Fatal(fmt.Errorf("can't save to %s: %w", path, err))
	}
	log.Print("Data was saved")
	return "Success", nil
}

func WriteToFileClientsCards(data []byte) (Result string, err error) {
	_ = os.Mkdir("backup", 0666)
	path := "backup/clientsCards.json"
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		log.Print("file \"", path, "\" not exist for write")
		log.Print("file \"", path, "\" will be create")
		var file, err = os.Create(path)
		if err != nil {
			return "file \"clientsCards.json\"could not be created successfully", err
		}
		defer file.Close()
		log.Print("file \"", path, "\" created successfully")
		log.Print("Data will saving")
		err = ioutil.WriteFile(path, data, 0666)
		if err != nil {
			log.Fatal(fmt.Errorf("can't save to %s: %w", path, err))
		}
		log.Print("Data was saved")
		return "Success", nil
	}
	log.Print("file ", path, " is exist")
	srcFile, err := os.Open(path)
	if err != nil {
		log.Fatal(fmt.Errorf("can't check source file %s: %w", path, err))
	}
	log.Print("source file was checked successfully")
	dateTimeBackup := time.Now().Format("backup/clientsCardsDataBackup(01-02-2006-15-04-5).json")
	log.Print("copy date and time for copy")
	fmt.Println(dateTimeBackup)
	log.Print("try create file to backup")
	destFile, err := os.Create(dateTimeBackup)
	if err != nil {
		log.Fatal(fmt.Errorf("can't create file to backup %s: %w", path, err))
	}
	log.Print("backup file was create")
	defer destFile.Close()
	log.Print("try copy from file to backup file")
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		log.Fatal(fmt.Errorf("can't finish copying %s: %w", path, err))
	}
	log.Print("file was copy to backup file")
	log.Print("Data will saving")
	err = ioutil.WriteFile(path, data, 0666)
	if err != nil {
		log.Fatal(fmt.Errorf("can't save to %s: %w", path, err))
	}
	log.Print("Data was saved")
	return "Success", nil
}

func WriteToFileATMs(data []byte) (Result string, err error) {
	_ = os.Mkdir("backup", 0666)
	path := "backup/atms.json"
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		log.Print("file \"", path, "\" not exist for write")
		log.Print("file \"", path, "\" will be create")
		var file, err = os.Create(path)
		if err != nil {
			return "file \"atms.json\" could not be created successfully", err
		}
		defer file.Close()
		log.Print("file \"", path, "\" created successfully")
		log.Print("Data will saving")
		err = ioutil.WriteFile(path, data, 0666)
		if err != nil {
			log.Fatal(fmt.Errorf("can't save to %s: %w", path, err))
		}
		log.Print("Data was saved")
		return "Success", nil
	}
	log.Print("file ", path, " is exist")
	srcFile, err := os.Open(path)
	if err != nil {
		log.Fatal(fmt.Errorf("can't check source file %s: %w", path, err))
	}
	log.Print("source file was checked successfully")
	dateTimeBackup := time.Now().Format("backup/atmsDataBackup(01-02-2006-15-04-5).json")
	log.Print("copy date and time for copy")
	fmt.Println(dateTimeBackup)
	log.Print("try create file to backup")
	destFile, err := os.Create(dateTimeBackup)
	if err != nil {
		log.Fatal(fmt.Errorf("can't create file to backup %s: %w", path, err))
	}
	log.Print("backup file was create")
	defer destFile.Close()
	log.Print("try copy from file to backup file")
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		log.Fatal(fmt.Errorf("can't finish copying %s: %w", path, err))
	}
	log.Print("file was copy to backup file")
	log.Print("Data will saving")
	err = ioutil.WriteFile(path, data, 0666)
	if err != nil {
		log.Fatal(fmt.Errorf("can't save to %s: %w", path, err))
	}
	log.Print("Data was saved")
	return "Success", nil
}

func WriteToFileServices(data []byte) (Result string, err error) {
	_ = os.Mkdir("backup", 0666)
	path := "backup/services.json"
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		log.Print("file \"", path, "\" not exist for write")
		log.Print("file \"", path, "\" will be create")
		var file, err = os.Create(path)
		if err != nil {
			return "file \"service.json\" could not be created successfully", err
		}
		defer file.Close()
		log.Print("file \"", path, "\" created successfully")
		log.Print("Data will saving")
		err = ioutil.WriteFile(path, data, 0666)
		if err != nil {
			log.Fatal(fmt.Errorf("can't save to %s: %w", path, err))
		}
		log.Print("Data was saved")
		return "Success", nil
	}
	log.Print("file ", path, " is exist")
	srcFile, err := os.Open(path)
	if err != nil {
		log.Fatal(fmt.Errorf("can't check source file %s: %w", path, err))
	}
	log.Print("source file was checked successfully")
	dateTimeBackup := time.Now().Format("backup/servicesDataBackup(01-02-2006-15-04-5).json")
	log.Print("copy date and time for copy")
	fmt.Println(dateTimeBackup)
	log.Print("try create file to backup")
	destFile, err := os.Create(dateTimeBackup)
	if err != nil {
		log.Fatal(fmt.Errorf("can't create file to backup %s: %w", path, err))
	}
	log.Print("backup file was create")
	defer destFile.Close()
	log.Print("try copy from file to backup file")
	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		log.Fatal(fmt.Errorf("can't finish copying %s: %w", path, err))
	}
	log.Print("file was copy to backup file")
	log.Print("Data will saving")
	err = ioutil.WriteFile(path, data, 0666)
	if err != nil {
		log.Fatal(fmt.Errorf("can't save to %s: %w", path, err))
	}
	log.Print("Data was saved")
	return "Success", nil
}

//TODO ############################### DO ALL FOR ME ###############################
func DoAllForMe(db *sql.DB) (Result string, err error) {
	managerStructuring, err := DbManagersToStruct(db)
	if err != nil {
		log.Fatalf("I can't converted Your sql manager data to your data []struct: %v", err)
	}
	managerByte, err := ManagersDataStructToBytes(managerStructuring)
	if err != nil {
		log.Fatalf("I can't converted Your manager data []struct to data []byte: %v", err)
	}
	Result, err = WriteToFileManagers(managerByte)
	if Result != "Success" {
		log.Fatalf("I can't write Your manager data []byte to file: it have error: %v", Result)
	}
	ClientStructuring, err := DbClientsToStruct(db)
	if err != nil {
		log.Fatalf("I can't converted Your sql client data to your data []struct: %v", err)
	}
	ClientByte, err := ClientDataStructToBytes(ClientStructuring)
	if err != nil {
		log.Fatalf("I can't converted Your client data []struct to data []byte: %v", err)
	}
	Result, err = WriteToFileClients(ClientByte)
	if Result != "Success" {
		log.Fatalf("I can't write Your client data []byte to file: it have error: %v", Result)
	}
	ClientCardStructuring, err := DbClientsCardsToStruct(db)
	if err != nil {
		log.Fatalf("I can't converted Your sql clientCard data to your data []struct: %v", err)
	}
	ClientCardByte, err := ClientsCardsDataStructToBytes(ClientCardStructuring)
	if err != nil {
		log.Fatalf("I can't converted Your clientCard data []struct to data []byte: %v", err)
	}
	Result, err = WriteToFileClientsCards(ClientCardByte)
	if Result != "Success" {
		log.Fatalf("I can't write Your clientCard data []byte to file: it have error: %v", Result)
	}
	ATMStructuring, err := DbATMsToStruct(db)
	if err != nil {
		log.Fatalf("I can't converted Your sql ATM data to your data []struct: %v", err)
	}
	ATMByte, err := ATMsDataStructToBytes(ATMStructuring)
	if err != nil {
		log.Fatalf("I can't converted Your ATM data []struct to data []byte: %v", err)
	}
	Result, err = WriteToFileATMs(ATMByte)
	if Result != "Success" {
		log.Fatalf("I can't write Your ATM data []byte to file: it have error: %v", Result)
	}
	ServiceStructuring, err := DbServicesToStruct(db)
	if err != nil {
		log.Fatalf("I can't converted Your sql service data to your data []struct: %v", err)
	}
	ServiceByte, err := ServicesDataStructToBytes(ServiceStructuring)
	if err != nil {
		log.Fatalf("I can't converted Your service data []struct to data []byte: %v", err)
	}
	Result, err = WriteToFileServices(ServiceByte)
	if Result != "Success" {
		log.Fatalf("I can't write Your service data []byte to file it have error: %v", Result)
	}
	return "YOU ARE LUCKY =)", nil
}

//TODO ############################### MARSHALING END HERE ###############################
