package core

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"
)

const (
	ConfPath string = "./src/Config/config.json"
	LogPath  string = "./src/Logs/"
)

type MongoDB struct {
	Mconn  *mgo.Session
	_db    string
	_table string
	_error bool
	_token string
}

func (this *MongoDB) Catch(str interface{}, proc ...bool) {

	newPath, exist := CreateFileLog(LogPath, "mdb-")

	if exist {

		f, err := os.OpenFile(newPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

		if err != nil {
			log.Println("Error en abrir el documento : "+LogPath, err)
		}
		defer f.Close()

		log.SetOutput(f)
	}

	if len(proc) > 0 {
		log.Println("--> TOKEN  : " + this._token)
		log.Println(str)
	} else {
		log.Println("CONNECTION  : [TOKEN]" + this._token)
		log.Println("CONNECTION :  [MONGODB] ")
		log.Println(str)
	}

}

func (this *MongoDB) GetConfig() (map[string]interface{}, bool) {

	data, err := ioutil.ReadFile(ConfPath)

	if err != nil {
		this.Catch(err)
		return nil, false
	}

	var resp map[string]interface{}

	if err := json.Unmarshal(data, &resp); err != nil {
		this.Catch(err)
		return nil, false
	}

	return resp, true
}

func (this *MongoDB) PrepareC(Credentials bool) (*mgo.Session, error) {

	conf, Cexist := this.GetConfig()
	var servers string = ""
	var err error
	var session *mgo.Session

	if Cexist == false {
		this.Catch("CONNECTION: " + "[FILE][EL ARCHIVO DE CONFIGURACION config.json NO SE ENCUENTRA EN LA DIRECCION {" + ConfPath + "} ]")
		return nil, errors.New("[FILE][EL ARCHIVO DE CONFIGURACION config.json NO SE ENCUENTRA EN LA DIRECCION {" + ConfPath + "} ]")
	}

	if Credentials == false {

		total := len(conf["servers"].([]interface{}))

		for _, value := range conf {
			if rec, ok := value.([]interface{}); ok {
				for key, val := range rec {
					servers += val.(string)
					if key != (total - 1) {
						servers += ","
					}
				}
			}
		}

		if servers == "" {
			this.Catch("CONNECTION: [CONFIG][NO EXISTEN SERVIDORES EN CONFIG.JSON ]")
			return nil, errors.New("CONNECTION: [CONFIG][NO EXISTEN SERVIDORES EN CONFIG.JSON ]")
		}

		session, err = mgo.Dial(servers)

	} else {

		this.Catch("CONNECTION: [FUNCTION][CREANDO CONEXION POR CREDENCIALES ]")

		total := len(conf["servers"].([]interface{}))
		Ostring := make([]string, total)

		for _, value := range conf {
			if rec, ok := value.([]interface{}); ok {
				for _, val := range rec {
					Ostring = append(Ostring, val.(string))
				}
			}

		}

		info := &mgo.DialInfo{
			Addrs:    Ostring,
			Username: conf["user"].(string),
			Password: conf["password"].(string),
			Database: conf["database"].(string),
			Timeout:  time.Second * 30,
			DialServer: func(addr *mgo.ServerAddr) (net.Conn, error) {
				this.Catch("--> ADDR :"+addr.String(), true)
				return tls.Dial("tcp", addr.String(), &tls.Config{})
			},
		}

		//this.Catch("HOST : ["+info.Addrs[1]+"]", true)
		this.Catch("-->	USUARIO  : ["+info.Username+"]", true)
		this.Catch("-->	PASSWORD  : [xxxxx...]", true)
		this.Catch("-->	DATABASE : ["+info.Database+"]", true)

		session, err = mgo.DialWithInfo(info)
	}

	t := int64(conf["timeout"].(float64))

	if err != nil {

		this.Catch("ERROR CAUSE : [" + err.Error() + "]")
		defer func() {
			if r := recover(); r == nil {
				this.Catch("ERROR CAUSE : [" + err.Error() + "]")
			}
		}()

	} else {
		this.Catch("--> ACTIVE : CONEXION ESTABLECIDA CON EXITO ", true)
		session.SetSocketTimeout(time.Hour * (time.Duration(t)))

	}

	return session, err

}

func (this *MongoDB) Conn(Credentials bool) {

	this._token = randomString(20)
	session, err := this.PrepareC(Credentials)

	if err != nil {
		this.Catch("CONNECTION: [ERROR][CONN NO SE PUEDE COMUNICAR]")
		this._error = true
	} else {
		this._error = false
		this.Mconn = session.Clone()
		this.Kill(false, session)
	}

}

func (this *MongoDB) Kill(all bool, session ...*mgo.Session) bool {

	if this._error == true {
		this.Catch("--> WARNING : NO HAY CONEXION A MATAR  ", true)
		return false
	}

	var sentinel bool = false

	if all == true {

		if session != nil {
			for _, op := range session {
				op.Close()
			}

			sentinel = true
		}

		if this.Mconn != nil {
			this.Mconn.Close()
			sentinel = true
		}

	} else {

		if session != nil {
			for _, op := range session {
				op.Close()
			}

			sentinel = true
		}

	}

	this.Catch("--> CONNECTION : CONEXION ELIMINADA CON EXITO   ", true)
	this.Catch("-------------------------------------------END OF LOCALE -----------------------------------------------------", true)
	return sentinel
}

func (this *MongoDB) EDatabase(database string, table string) (string, string) {

	var a, b string = "", ""

	if database != "" {
		a = database
	} else {
		a = this._db
	}

	if table != "" {
		b = table
	} else {
		b = this._table
	}

	return a, b
}

func (this *MongoDB) GetSession() *mgo.Session {
	return this.Mconn
}

func (this *MongoDB) InstanceDB(Dbname string) {
	this._db = Dbname
}

func (this *MongoDB) InstanceTable(Table string) {
	this._table = Table
}

func (this *MongoDB) FindBy(pointer interface{}, query bson.M, params ...string) (interface{}, error) {

	var flog string

	if this._error == true {

		flog = "---> FindBy : [No se pudo ejecutar debido a que la conexion fue rechazada]"
		flog += "\t\t\n --> PARAM 1 :" + params[0]
		flog += "\t\t\n --> PARAM 2 :" + params[1]
		this.Catch(flog+"\n\n", true)

		return nil, errors.New("_")
	}

	d, t := this.EDatabase(params[0], params[1])
	c := this.Mconn.DB(d).C(t)
	err := c.Find(query).One(&pointer)

	return pointer, err
}

func (this *MongoDB) FindAll(pointer interface{}, query bson.M, params ...string) (interface{}, error) {

	if this._error == true {
		return nil, errors.New("_")
	}

	d, t := this.EDatabase(params[0], params[1])
	c := this.Mconn.DB(d).C(t)
	e := c.Find(query).All(&pointer)
	return pointer, e
}

func (this *MongoDB) FindAndPaginate(pointer interface{}, query bson.M, start int, end int, params ...string) (interface{}, error) {

	if this._error == true {
		return nil, errors.New("_")
	}

	d, t := this.EDatabase(params[0], params[1])
	c := this.Mconn.DB(d).C(t)
	e := c.Find(query).Skip(start).Limit(end).All(&pointer)
	return pointer, e
}

func (this *MongoDB) Count(query bson.M, params ...string) (int, error) {

	if this._error == true {
		return -1, errors.New("_")
	}

	d, t := this.EDatabase(params[0], params[1])
	c := this.Mconn.DB(d).C(t)
	r, err := c.Find(query).Count()
	return r, err
}

func (this *MongoDB) CreateCollection() {

}
