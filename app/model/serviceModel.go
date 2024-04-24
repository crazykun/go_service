package model

type ServiceModel struct {
	Id         int64  `json:"id"`
	Name       string `json:"name"`
	Title      string `json:"title"`
	Dir        string `json:"dir"`
	CmdStart   string `json:"cmd_start"`
	CmdStop    string `json:"cmd_stop"`
	CmdRestart string `json:"cmd_restart"`
	Port       int64  `json:"port"`
	Remark     string `json:"remark"`
}

type ServiceStatusModel struct {
	ServiceModel
	Status  int    `json:"status"`
	Pid     string `json:"pid"`
	Process string `json:"process"`
}

func (s ServiceModel) TableName() string {
	return "service"
}
