# atc_thrift
--
    import "."


## Usage

```go
var Corruption_Correlation_DEFAULT float64 = 0
```

```go
var Delay_Correlation_DEFAULT float64 = 0
```

```go
var Delay_Jitter_DEFAULT int32 = 0
```

```go
var GetGroupTokenResult_Success_DEFAULT string
```

```go
var GoUnusedProtection__ int
```

```go
var Loss_Correlation_DEFAULT float64 = 0
```

```go
var Reorder_Correlation_DEFAULT float64 = 0
```

```go
var Shaping_IptablesOptions_DEFAULT []string
```

#### type Atcd

```go
type Atcd interface {
	GetAtcdInfo() (r *AtcdInfo, err error)
	// Parameters:
	//  - Member
	CreateGroup(member string) (r *ShapingGroup, err error)
	// Parameters:
	//  - Id
	GetGroup(id int64) (r *ShapingGroup, err error)
	// Parameters:
	//  - Member
	GetGroupWith(member string) (r *ShapingGroup, err error)
	// Parameters:
	//  - Id
	GetGroupToken(id int64) (r string, err error)
	// Parameters:
	//  - Id
	//  - ToRemove
	//  - Token
	LeaveGroup(id int64, to_remove string, token string) (err error)
	// Parameters:
	//  - Id
	//  - ToAdd
	//  - Token
	JoinGroup(id int64, to_add string, token string) (err error)
	// Parameters:
	//  - Id
	//  - Settings
	//  - Token
	ShapeGroup(id int64, settings *Setting, token string) (r *Setting, err error)
	// Parameters:
	//  - Id
	//  - Token
	UnshapeGroup(id int64, token string) (err error)
}
```


#### type AtcdClient

```go
type AtcdClient struct {
	Transport       thrift.TTransport
	ProtocolFactory thrift.TProtocolFactory
	InputProtocol   thrift.TProtocol
	OutputProtocol  thrift.TProtocol
	SeqId           int32
}
```


#### func  NewAtcdClientFactory

```go
func NewAtcdClientFactory(t thrift.TTransport, f thrift.TProtocolFactory) *AtcdClient
```

#### func  NewAtcdClientProtocol

```go
func NewAtcdClientProtocol(t thrift.TTransport, iprot thrift.TProtocol, oprot thrift.TProtocol) *AtcdClient
```

#### func (*AtcdClient) CreateGroup

```go
func (p *AtcdClient) CreateGroup(member string) (r *ShapingGroup, err error)
```
Parameters:

    - Member

#### func (*AtcdClient) GetAtcdInfo

```go
func (p *AtcdClient) GetAtcdInfo() (r *AtcdInfo, err error)
```

#### func (*AtcdClient) GetGroup

```go
func (p *AtcdClient) GetGroup(id int64) (r *ShapingGroup, err error)
```
Parameters:

    - Id

#### func (*AtcdClient) GetGroupToken

```go
func (p *AtcdClient) GetGroupToken(id int64) (r string, err error)
```
Parameters:

    - Id

#### func (*AtcdClient) GetGroupWith

```go
func (p *AtcdClient) GetGroupWith(member string) (r *ShapingGroup, err error)
```
Parameters:

    - Member

#### func (*AtcdClient) JoinGroup

```go
func (p *AtcdClient) JoinGroup(id int64, to_add string, token string) (err error)
```
Parameters:

    - Id
    - ToAdd
    - Token

#### func (*AtcdClient) LeaveGroup

```go
func (p *AtcdClient) LeaveGroup(id int64, to_remove string, token string) (err error)
```
Parameters:

    - Id
    - ToRemove
    - Token

#### func (*AtcdClient) ShapeGroup

```go
func (p *AtcdClient) ShapeGroup(id int64, settings *Setting, token string) (r *Setting, err error)
```
Parameters:

    - Id
    - Settings
    - Token

#### func (*AtcdClient) UnshapeGroup

```go
func (p *AtcdClient) UnshapeGroup(id int64, token string) (err error)
```
Parameters:

    - Id
    - Token

#### type AtcdInfo

```go
type AtcdInfo struct {
	Platform PlatformType `thrift:"platform,1" json:"platform"`
	Version  string       `thrift:"version,2" json:"version"`
}
```


```go
var GetAtcdInfoResult_Success_DEFAULT *AtcdInfo
```

#### func  NewAtcdInfo

```go
func NewAtcdInfo() *AtcdInfo
```

#### func (*AtcdInfo) GetPlatform

```go
func (p *AtcdInfo) GetPlatform() PlatformType
```

#### func (*AtcdInfo) GetVersion

```go
func (p *AtcdInfo) GetVersion() string
```

#### func (*AtcdInfo) Read

```go
func (p *AtcdInfo) Read(iprot thrift.TProtocol) error
```

#### func (*AtcdInfo) ReadField1

```go
func (p *AtcdInfo) ReadField1(iprot thrift.TProtocol) error
```

#### func (*AtcdInfo) ReadField2

```go
func (p *AtcdInfo) ReadField2(iprot thrift.TProtocol) error
```

#### func (*AtcdInfo) String

```go
func (p *AtcdInfo) String() string
```

#### func (*AtcdInfo) Write

```go
func (p *AtcdInfo) Write(oprot thrift.TProtocol) error
```

#### type AtcdProcessor

```go
type AtcdProcessor struct {
}
```


#### func  NewAtcdProcessor

```go
func NewAtcdProcessor(handler Atcd) *AtcdProcessor
```

#### func (*AtcdProcessor) AddToProcessorMap

```go
func (p *AtcdProcessor) AddToProcessorMap(key string, processor thrift.TProcessorFunction)
```

#### func (*AtcdProcessor) GetProcessorFunction

```go
func (p *AtcdProcessor) GetProcessorFunction(key string) (processor thrift.TProcessorFunction, ok bool)
```

#### func (*AtcdProcessor) Process

```go
func (p *AtcdProcessor) Process(iprot, oprot thrift.TProtocol) (success bool, err thrift.TException)
```

#### func (*AtcdProcessor) ProcessorMap

```go
func (p *AtcdProcessor) ProcessorMap() map[string]thrift.TProcessorFunction
```

#### type Corruption

```go
type Corruption struct {
	Percentage  float64 `thrift:"percentage,1" json:"percentage"`
	Correlation float64 `thrift:"correlation,2" json:"correlation"`
}
```


```go
var Shaping_Corruption_DEFAULT *Corruption = &Corruption{
	Percentage: 0}
```

#### func  NewCorruption

```go
func NewCorruption() *Corruption
```

#### func (*Corruption) GetCorrelation

```go
func (p *Corruption) GetCorrelation() float64
```

#### func (*Corruption) GetPercentage

```go
func (p *Corruption) GetPercentage() float64
```

#### func (*Corruption) IsSetCorrelation

```go
func (p *Corruption) IsSetCorrelation() bool
```

#### func (*Corruption) Read

```go
func (p *Corruption) Read(iprot thrift.TProtocol) error
```

#### func (*Corruption) ReadField1

```go
func (p *Corruption) ReadField1(iprot thrift.TProtocol) error
```

#### func (*Corruption) ReadField2

```go
func (p *Corruption) ReadField2(iprot thrift.TProtocol) error
```

#### func (*Corruption) String

```go
func (p *Corruption) String() string
```

#### func (*Corruption) Write

```go
func (p *Corruption) Write(oprot thrift.TProtocol) error
```

#### type CreateGroupArgs

```go
type CreateGroupArgs struct {
	Member string `thrift:"member,1" json:"member"`
}
```


#### func  NewCreateGroupArgs

```go
func NewCreateGroupArgs() *CreateGroupArgs
```

#### func (*CreateGroupArgs) GetMember

```go
func (p *CreateGroupArgs) GetMember() string
```

#### func (*CreateGroupArgs) Read

```go
func (p *CreateGroupArgs) Read(iprot thrift.TProtocol) error
```

#### func (*CreateGroupArgs) ReadField1

```go
func (p *CreateGroupArgs) ReadField1(iprot thrift.TProtocol) error
```

#### func (*CreateGroupArgs) String

```go
func (p *CreateGroupArgs) String() string
```

#### func (*CreateGroupArgs) Write

```go
func (p *CreateGroupArgs) Write(oprot thrift.TProtocol) error
```

#### type CreateGroupResult

```go
type CreateGroupResult struct {
	Success *ShapingGroup `thrift:"success,0" json:"success"`
}
```


#### func  NewCreateGroupResult

```go
func NewCreateGroupResult() *CreateGroupResult
```

#### func (*CreateGroupResult) GetSuccess

```go
func (p *CreateGroupResult) GetSuccess() *ShapingGroup
```

#### func (*CreateGroupResult) IsSetSuccess

```go
func (p *CreateGroupResult) IsSetSuccess() bool
```

#### func (*CreateGroupResult) Read

```go
func (p *CreateGroupResult) Read(iprot thrift.TProtocol) error
```

#### func (*CreateGroupResult) ReadField0

```go
func (p *CreateGroupResult) ReadField0(iprot thrift.TProtocol) error
```

#### func (*CreateGroupResult) String

```go
func (p *CreateGroupResult) String() string
```

#### func (*CreateGroupResult) Write

```go
func (p *CreateGroupResult) Write(oprot thrift.TProtocol) error
```

#### type Delay

```go
type Delay struct {
	Delay       int32   `thrift:"delay,1" json:"delay"`
	Jitter      int32   `thrift:"jitter,2" json:"jitter"`
	Correlation float64 `thrift:"correlation,3" json:"correlation"`
}
```


```go
var Shaping_Delay_DEFAULT *Delay = &Delay{
	Delay: 0}
```

#### func  NewDelay

```go
func NewDelay() *Delay
```

#### func (*Delay) GetCorrelation

```go
func (p *Delay) GetCorrelation() float64
```

#### func (*Delay) GetDelay

```go
func (p *Delay) GetDelay() int32
```

#### func (*Delay) GetJitter

```go
func (p *Delay) GetJitter() int32
```

#### func (*Delay) IsSetCorrelation

```go
func (p *Delay) IsSetCorrelation() bool
```

#### func (*Delay) IsSetJitter

```go
func (p *Delay) IsSetJitter() bool
```

#### func (*Delay) Read

```go
func (p *Delay) Read(iprot thrift.TProtocol) error
```

#### func (*Delay) ReadField1

```go
func (p *Delay) ReadField1(iprot thrift.TProtocol) error
```

#### func (*Delay) ReadField2

```go
func (p *Delay) ReadField2(iprot thrift.TProtocol) error
```

#### func (*Delay) ReadField3

```go
func (p *Delay) ReadField3(iprot thrift.TProtocol) error
```

#### func (*Delay) String

```go
func (p *Delay) String() string
```

#### func (*Delay) Write

```go
func (p *Delay) Write(oprot thrift.TProtocol) error
```

#### type GetAtcdInfoArgs

```go
type GetAtcdInfoArgs struct {
}
```


#### func  NewGetAtcdInfoArgs

```go
func NewGetAtcdInfoArgs() *GetAtcdInfoArgs
```

#### func (*GetAtcdInfoArgs) Read

```go
func (p *GetAtcdInfoArgs) Read(iprot thrift.TProtocol) error
```

#### func (*GetAtcdInfoArgs) String

```go
func (p *GetAtcdInfoArgs) String() string
```

#### func (*GetAtcdInfoArgs) Write

```go
func (p *GetAtcdInfoArgs) Write(oprot thrift.TProtocol) error
```

#### type GetAtcdInfoResult

```go
type GetAtcdInfoResult struct {
	Success *AtcdInfo `thrift:"success,0" json:"success"`
}
```


#### func  NewGetAtcdInfoResult

```go
func NewGetAtcdInfoResult() *GetAtcdInfoResult
```

#### func (*GetAtcdInfoResult) GetSuccess

```go
func (p *GetAtcdInfoResult) GetSuccess() *AtcdInfo
```

#### func (*GetAtcdInfoResult) IsSetSuccess

```go
func (p *GetAtcdInfoResult) IsSetSuccess() bool
```

#### func (*GetAtcdInfoResult) Read

```go
func (p *GetAtcdInfoResult) Read(iprot thrift.TProtocol) error
```

#### func (*GetAtcdInfoResult) ReadField0

```go
func (p *GetAtcdInfoResult) ReadField0(iprot thrift.TProtocol) error
```

#### func (*GetAtcdInfoResult) String

```go
func (p *GetAtcdInfoResult) String() string
```

#### func (*GetAtcdInfoResult) Write

```go
func (p *GetAtcdInfoResult) Write(oprot thrift.TProtocol) error
```

#### type GetGroupArgs

```go
type GetGroupArgs struct {
	Id int64 `thrift:"id,1" json:"id"`
}
```


#### func  NewGetGroupArgs

```go
func NewGetGroupArgs() *GetGroupArgs
```

#### func (*GetGroupArgs) GetId

```go
func (p *GetGroupArgs) GetId() int64
```

#### func (*GetGroupArgs) Read

```go
func (p *GetGroupArgs) Read(iprot thrift.TProtocol) error
```

#### func (*GetGroupArgs) ReadField1

```go
func (p *GetGroupArgs) ReadField1(iprot thrift.TProtocol) error
```

#### func (*GetGroupArgs) String

```go
func (p *GetGroupArgs) String() string
```

#### func (*GetGroupArgs) Write

```go
func (p *GetGroupArgs) Write(oprot thrift.TProtocol) error
```

#### type GetGroupResult

```go
type GetGroupResult struct {
	Success *ShapingGroup `thrift:"success,0" json:"success"`
}
```


#### func  NewGetGroupResult

```go
func NewGetGroupResult() *GetGroupResult
```

#### func (*GetGroupResult) GetSuccess

```go
func (p *GetGroupResult) GetSuccess() *ShapingGroup
```

#### func (*GetGroupResult) IsSetSuccess

```go
func (p *GetGroupResult) IsSetSuccess() bool
```

#### func (*GetGroupResult) Read

```go
func (p *GetGroupResult) Read(iprot thrift.TProtocol) error
```

#### func (*GetGroupResult) ReadField0

```go
func (p *GetGroupResult) ReadField0(iprot thrift.TProtocol) error
```

#### func (*GetGroupResult) String

```go
func (p *GetGroupResult) String() string
```

#### func (*GetGroupResult) Write

```go
func (p *GetGroupResult) Write(oprot thrift.TProtocol) error
```

#### type GetGroupTokenArgs

```go
type GetGroupTokenArgs struct {
	Id int64 `thrift:"id,1" json:"id"`
}
```


#### func  NewGetGroupTokenArgs

```go
func NewGetGroupTokenArgs() *GetGroupTokenArgs
```

#### func (*GetGroupTokenArgs) GetId

```go
func (p *GetGroupTokenArgs) GetId() int64
```

#### func (*GetGroupTokenArgs) Read

```go
func (p *GetGroupTokenArgs) Read(iprot thrift.TProtocol) error
```

#### func (*GetGroupTokenArgs) ReadField1

```go
func (p *GetGroupTokenArgs) ReadField1(iprot thrift.TProtocol) error
```

#### func (*GetGroupTokenArgs) String

```go
func (p *GetGroupTokenArgs) String() string
```

#### func (*GetGroupTokenArgs) Write

```go
func (p *GetGroupTokenArgs) Write(oprot thrift.TProtocol) error
```

#### type GetGroupTokenResult

```go
type GetGroupTokenResult struct {
	Success *string `thrift:"success,0" json:"success"`
}
```


#### func  NewGetGroupTokenResult

```go
func NewGetGroupTokenResult() *GetGroupTokenResult
```

#### func (*GetGroupTokenResult) GetSuccess

```go
func (p *GetGroupTokenResult) GetSuccess() string
```

#### func (*GetGroupTokenResult) IsSetSuccess

```go
func (p *GetGroupTokenResult) IsSetSuccess() bool
```

#### func (*GetGroupTokenResult) Read

```go
func (p *GetGroupTokenResult) Read(iprot thrift.TProtocol) error
```

#### func (*GetGroupTokenResult) ReadField0

```go
func (p *GetGroupTokenResult) ReadField0(iprot thrift.TProtocol) error
```

#### func (*GetGroupTokenResult) String

```go
func (p *GetGroupTokenResult) String() string
```

#### func (*GetGroupTokenResult) Write

```go
func (p *GetGroupTokenResult) Write(oprot thrift.TProtocol) error
```

#### type GetGroupWithArgs

```go
type GetGroupWithArgs struct {
	Member string `thrift:"member,1" json:"member"`
}
```


#### func  NewGetGroupWithArgs

```go
func NewGetGroupWithArgs() *GetGroupWithArgs
```

#### func (*GetGroupWithArgs) GetMember

```go
func (p *GetGroupWithArgs) GetMember() string
```

#### func (*GetGroupWithArgs) Read

```go
func (p *GetGroupWithArgs) Read(iprot thrift.TProtocol) error
```

#### func (*GetGroupWithArgs) ReadField1

```go
func (p *GetGroupWithArgs) ReadField1(iprot thrift.TProtocol) error
```

#### func (*GetGroupWithArgs) String

```go
func (p *GetGroupWithArgs) String() string
```

#### func (*GetGroupWithArgs) Write

```go
func (p *GetGroupWithArgs) Write(oprot thrift.TProtocol) error
```

#### type GetGroupWithResult

```go
type GetGroupWithResult struct {
	Success *ShapingGroup `thrift:"success,0" json:"success"`
}
```


#### func  NewGetGroupWithResult

```go
func NewGetGroupWithResult() *GetGroupWithResult
```

#### func (*GetGroupWithResult) GetSuccess

```go
func (p *GetGroupWithResult) GetSuccess() *ShapingGroup
```

#### func (*GetGroupWithResult) IsSetSuccess

```go
func (p *GetGroupWithResult) IsSetSuccess() bool
```

#### func (*GetGroupWithResult) Read

```go
func (p *GetGroupWithResult) Read(iprot thrift.TProtocol) error
```

#### func (*GetGroupWithResult) ReadField0

```go
func (p *GetGroupWithResult) ReadField0(iprot thrift.TProtocol) error
```

#### func (*GetGroupWithResult) String

```go
func (p *GetGroupWithResult) String() string
```

#### func (*GetGroupWithResult) Write

```go
func (p *GetGroupWithResult) Write(oprot thrift.TProtocol) error
```

#### type JoinGroupArgs

```go
type JoinGroupArgs struct {
	Id    int64  `thrift:"id,1" json:"id"`
	ToAdd string `thrift:"to_add,2" json:"to_add"`
	Token string `thrift:"token,3" json:"token"`
}
```


#### func  NewJoinGroupArgs

```go
func NewJoinGroupArgs() *JoinGroupArgs
```

#### func (*JoinGroupArgs) GetId

```go
func (p *JoinGroupArgs) GetId() int64
```

#### func (*JoinGroupArgs) GetToAdd

```go
func (p *JoinGroupArgs) GetToAdd() string
```

#### func (*JoinGroupArgs) GetToken

```go
func (p *JoinGroupArgs) GetToken() string
```

#### func (*JoinGroupArgs) Read

```go
func (p *JoinGroupArgs) Read(iprot thrift.TProtocol) error
```

#### func (*JoinGroupArgs) ReadField1

```go
func (p *JoinGroupArgs) ReadField1(iprot thrift.TProtocol) error
```

#### func (*JoinGroupArgs) ReadField2

```go
func (p *JoinGroupArgs) ReadField2(iprot thrift.TProtocol) error
```

#### func (*JoinGroupArgs) ReadField3

```go
func (p *JoinGroupArgs) ReadField3(iprot thrift.TProtocol) error
```

#### func (*JoinGroupArgs) String

```go
func (p *JoinGroupArgs) String() string
```

#### func (*JoinGroupArgs) Write

```go
func (p *JoinGroupArgs) Write(oprot thrift.TProtocol) error
```

#### type JoinGroupResult

```go
type JoinGroupResult struct {
}
```


#### func  NewJoinGroupResult

```go
func NewJoinGroupResult() *JoinGroupResult
```

#### func (*JoinGroupResult) Read

```go
func (p *JoinGroupResult) Read(iprot thrift.TProtocol) error
```

#### func (*JoinGroupResult) String

```go
func (p *JoinGroupResult) String() string
```

#### func (*JoinGroupResult) Write

```go
func (p *JoinGroupResult) Write(oprot thrift.TProtocol) error
```

#### type LeaveGroupArgs

```go
type LeaveGroupArgs struct {
	Id       int64  `thrift:"id,1" json:"id"`
	ToRemove string `thrift:"to_remove,2" json:"to_remove"`
	Token    string `thrift:"token,3" json:"token"`
}
```


#### func  NewLeaveGroupArgs

```go
func NewLeaveGroupArgs() *LeaveGroupArgs
```

#### func (*LeaveGroupArgs) GetId

```go
func (p *LeaveGroupArgs) GetId() int64
```

#### func (*LeaveGroupArgs) GetToRemove

```go
func (p *LeaveGroupArgs) GetToRemove() string
```

#### func (*LeaveGroupArgs) GetToken

```go
func (p *LeaveGroupArgs) GetToken() string
```

#### func (*LeaveGroupArgs) Read

```go
func (p *LeaveGroupArgs) Read(iprot thrift.TProtocol) error
```

#### func (*LeaveGroupArgs) ReadField1

```go
func (p *LeaveGroupArgs) ReadField1(iprot thrift.TProtocol) error
```

#### func (*LeaveGroupArgs) ReadField2

```go
func (p *LeaveGroupArgs) ReadField2(iprot thrift.TProtocol) error
```

#### func (*LeaveGroupArgs) ReadField3

```go
func (p *LeaveGroupArgs) ReadField3(iprot thrift.TProtocol) error
```

#### func (*LeaveGroupArgs) String

```go
func (p *LeaveGroupArgs) String() string
```

#### func (*LeaveGroupArgs) Write

```go
func (p *LeaveGroupArgs) Write(oprot thrift.TProtocol) error
```

#### type LeaveGroupResult

```go
type LeaveGroupResult struct {
}
```


#### func  NewLeaveGroupResult

```go
func NewLeaveGroupResult() *LeaveGroupResult
```

#### func (*LeaveGroupResult) Read

```go
func (p *LeaveGroupResult) Read(iprot thrift.TProtocol) error
```

#### func (*LeaveGroupResult) String

```go
func (p *LeaveGroupResult) String() string
```

#### func (*LeaveGroupResult) Write

```go
func (p *LeaveGroupResult) Write(oprot thrift.TProtocol) error
```

#### type Loss

```go
type Loss struct {
	Percentage  float64 `thrift:"percentage,1" json:"percentage"`
	Correlation float64 `thrift:"correlation,2" json:"correlation"`
}
```


```go
var Shaping_Loss_DEFAULT *Loss = &Loss{
	Percentage: 0}
```

#### func  NewLoss

```go
func NewLoss() *Loss
```

#### func (*Loss) GetCorrelation

```go
func (p *Loss) GetCorrelation() float64
```

#### func (*Loss) GetPercentage

```go
func (p *Loss) GetPercentage() float64
```

#### func (*Loss) IsSetCorrelation

```go
func (p *Loss) IsSetCorrelation() bool
```

#### func (*Loss) Read

```go
func (p *Loss) Read(iprot thrift.TProtocol) error
```

#### func (*Loss) ReadField1

```go
func (p *Loss) ReadField1(iprot thrift.TProtocol) error
```

#### func (*Loss) ReadField2

```go
func (p *Loss) ReadField2(iprot thrift.TProtocol) error
```

#### func (*Loss) String

```go
func (p *Loss) String() string
```

#### func (*Loss) Write

```go
func (p *Loss) Write(oprot thrift.TProtocol) error
```

#### type PlatformType

```go
type PlatformType int64
```


```go
const (
	PlatformType_LINUX PlatformType = 0
)
```

#### func  PlatformTypeFromString

```go
func PlatformTypeFromString(s string) (PlatformType, error)
```

#### func  PlatformTypePtr

```go
func PlatformTypePtr(v PlatformType) *PlatformType
```

#### func (PlatformType) String

```go
func (p PlatformType) String() string
```

#### type Reorder

```go
type Reorder struct {
	Percentage  float64 `thrift:"percentage,1" json:"percentage"`
	Gap         int32   `thrift:"gap,2" json:"gap"`
	Correlation float64 `thrift:"correlation,3" json:"correlation"`
}
```


```go
var Shaping_Reorder_DEFAULT *Reorder = &Reorder{
	Percentage: 0}
```

#### func  NewReorder

```go
func NewReorder() *Reorder
```

#### func (*Reorder) GetCorrelation

```go
func (p *Reorder) GetCorrelation() float64
```

#### func (*Reorder) GetGap

```go
func (p *Reorder) GetGap() int32
```

#### func (*Reorder) GetPercentage

```go
func (p *Reorder) GetPercentage() float64
```

#### func (*Reorder) IsSetCorrelation

```go
func (p *Reorder) IsSetCorrelation() bool
```

#### func (*Reorder) Read

```go
func (p *Reorder) Read(iprot thrift.TProtocol) error
```

#### func (*Reorder) ReadField1

```go
func (p *Reorder) ReadField1(iprot thrift.TProtocol) error
```

#### func (*Reorder) ReadField2

```go
func (p *Reorder) ReadField2(iprot thrift.TProtocol) error
```

#### func (*Reorder) ReadField3

```go
func (p *Reorder) ReadField3(iprot thrift.TProtocol) error
```

#### func (*Reorder) String

```go
func (p *Reorder) String() string
```

#### func (*Reorder) Write

```go
func (p *Reorder) Write(oprot thrift.TProtocol) error
```

#### type Setting

```go
type Setting struct {
	Up   *Shaping `thrift:"up,1" json:"up"`
	Down *Shaping `thrift:"down,2" json:"down"`
}
```


```go
var ShapeGroupArgs_Settings_DEFAULT *Setting
```

```go
var ShapeGroupResult_Success_DEFAULT *Setting
```

```go
var ShapingGroup_Shaping_DEFAULT *Setting
```

#### func  NewSetting

```go
func NewSetting() *Setting
```

#### func (*Setting) GetDown

```go
func (p *Setting) GetDown() *Shaping
```

#### func (*Setting) GetUp

```go
func (p *Setting) GetUp() *Shaping
```

#### func (*Setting) IsSetDown

```go
func (p *Setting) IsSetDown() bool
```

#### func (*Setting) IsSetUp

```go
func (p *Setting) IsSetUp() bool
```

#### func (*Setting) Read

```go
func (p *Setting) Read(iprot thrift.TProtocol) error
```

#### func (*Setting) ReadField1

```go
func (p *Setting) ReadField1(iprot thrift.TProtocol) error
```

#### func (*Setting) ReadField2

```go
func (p *Setting) ReadField2(iprot thrift.TProtocol) error
```

#### func (*Setting) String

```go
func (p *Setting) String() string
```

#### func (*Setting) Write

```go
func (p *Setting) Write(oprot thrift.TProtocol) error
```

#### type ShapeGroupArgs

```go
type ShapeGroupArgs struct {
	Id       int64    `thrift:"id,1" json:"id"`
	Settings *Setting `thrift:"settings,2" json:"settings"`
	Token    string   `thrift:"token,3" json:"token"`
}
```


#### func  NewShapeGroupArgs

```go
func NewShapeGroupArgs() *ShapeGroupArgs
```

#### func (*ShapeGroupArgs) GetId

```go
func (p *ShapeGroupArgs) GetId() int64
```

#### func (*ShapeGroupArgs) GetSettings

```go
func (p *ShapeGroupArgs) GetSettings() *Setting
```

#### func (*ShapeGroupArgs) GetToken

```go
func (p *ShapeGroupArgs) GetToken() string
```

#### func (*ShapeGroupArgs) IsSetSettings

```go
func (p *ShapeGroupArgs) IsSetSettings() bool
```

#### func (*ShapeGroupArgs) Read

```go
func (p *ShapeGroupArgs) Read(iprot thrift.TProtocol) error
```

#### func (*ShapeGroupArgs) ReadField1

```go
func (p *ShapeGroupArgs) ReadField1(iprot thrift.TProtocol) error
```

#### func (*ShapeGroupArgs) ReadField2

```go
func (p *ShapeGroupArgs) ReadField2(iprot thrift.TProtocol) error
```

#### func (*ShapeGroupArgs) ReadField3

```go
func (p *ShapeGroupArgs) ReadField3(iprot thrift.TProtocol) error
```

#### func (*ShapeGroupArgs) String

```go
func (p *ShapeGroupArgs) String() string
```

#### func (*ShapeGroupArgs) Write

```go
func (p *ShapeGroupArgs) Write(oprot thrift.TProtocol) error
```

#### type ShapeGroupResult

```go
type ShapeGroupResult struct {
	Success *Setting `thrift:"success,0" json:"success"`
}
```


#### func  NewShapeGroupResult

```go
func NewShapeGroupResult() *ShapeGroupResult
```

#### func (*ShapeGroupResult) GetSuccess

```go
func (p *ShapeGroupResult) GetSuccess() *Setting
```

#### func (*ShapeGroupResult) IsSetSuccess

```go
func (p *ShapeGroupResult) IsSetSuccess() bool
```

#### func (*ShapeGroupResult) Read

```go
func (p *ShapeGroupResult) Read(iprot thrift.TProtocol) error
```

#### func (*ShapeGroupResult) ReadField0

```go
func (p *ShapeGroupResult) ReadField0(iprot thrift.TProtocol) error
```

#### func (*ShapeGroupResult) String

```go
func (p *ShapeGroupResult) String() string
```

#### func (*ShapeGroupResult) Write

```go
func (p *ShapeGroupResult) Write(oprot thrift.TProtocol) error
```

#### type Shaping

```go
type Shaping struct {
	Rate            int32       `thrift:"rate,1" json:"rate"`
	Delay           *Delay      `thrift:"delay,2" json:"delay"`
	Loss            *Loss       `thrift:"loss,3" json:"loss"`
	Reorder         *Reorder    `thrift:"reorder,4" json:"reorder"`
	Corruption      *Corruption `thrift:"corruption,5" json:"corruption"`
	IptablesOptions []string    `thrift:"iptables_options,6" json:"iptables_options"`
}
```


```go
var Setting_Down_DEFAULT *Shaping
```

```go
var Setting_Up_DEFAULT *Shaping
```

#### func  NewShaping

```go
func NewShaping() *Shaping
```

#### func (*Shaping) GetCorruption

```go
func (p *Shaping) GetCorruption() *Corruption
```

#### func (*Shaping) GetDelay

```go
func (p *Shaping) GetDelay() *Delay
```

#### func (*Shaping) GetIptablesOptions

```go
func (p *Shaping) GetIptablesOptions() []string
```

#### func (*Shaping) GetLoss

```go
func (p *Shaping) GetLoss() *Loss
```

#### func (*Shaping) GetRate

```go
func (p *Shaping) GetRate() int32
```

#### func (*Shaping) GetReorder

```go
func (p *Shaping) GetReorder() *Reorder
```

#### func (*Shaping) IsSetCorruption

```go
func (p *Shaping) IsSetCorruption() bool
```

#### func (*Shaping) IsSetDelay

```go
func (p *Shaping) IsSetDelay() bool
```

#### func (*Shaping) IsSetIptablesOptions

```go
func (p *Shaping) IsSetIptablesOptions() bool
```

#### func (*Shaping) IsSetLoss

```go
func (p *Shaping) IsSetLoss() bool
```

#### func (*Shaping) IsSetReorder

```go
func (p *Shaping) IsSetReorder() bool
```

#### func (*Shaping) Read

```go
func (p *Shaping) Read(iprot thrift.TProtocol) error
```

#### func (*Shaping) ReadField1

```go
func (p *Shaping) ReadField1(iprot thrift.TProtocol) error
```

#### func (*Shaping) ReadField2

```go
func (p *Shaping) ReadField2(iprot thrift.TProtocol) error
```

#### func (*Shaping) ReadField3

```go
func (p *Shaping) ReadField3(iprot thrift.TProtocol) error
```

#### func (*Shaping) ReadField4

```go
func (p *Shaping) ReadField4(iprot thrift.TProtocol) error
```

#### func (*Shaping) ReadField5

```go
func (p *Shaping) ReadField5(iprot thrift.TProtocol) error
```

#### func (*Shaping) ReadField6

```go
func (p *Shaping) ReadField6(iprot thrift.TProtocol) error
```

#### func (*Shaping) String

```go
func (p *Shaping) String() string
```

#### func (*Shaping) Write

```go
func (p *Shaping) Write(oprot thrift.TProtocol) error
```

#### type ShapingGroup

```go
type ShapingGroup struct {
	Id      int64    `thrift:"id,1" json:"id"`
	Members []string `thrift:"members,2" json:"members"`
	Shaping *Setting `thrift:"shaping,3" json:"shaping"`
}
```


```go
var CreateGroupResult_Success_DEFAULT *ShapingGroup
```

```go
var GetGroupResult_Success_DEFAULT *ShapingGroup
```

```go
var GetGroupWithResult_Success_DEFAULT *ShapingGroup
```

#### func  NewShapingGroup

```go
func NewShapingGroup() *ShapingGroup
```

#### func (*ShapingGroup) GetId

```go
func (p *ShapingGroup) GetId() int64
```

#### func (*ShapingGroup) GetMembers

```go
func (p *ShapingGroup) GetMembers() []string
```

#### func (*ShapingGroup) GetShaping

```go
func (p *ShapingGroup) GetShaping() *Setting
```

#### func (*ShapingGroup) IsSetShaping

```go
func (p *ShapingGroup) IsSetShaping() bool
```

#### func (*ShapingGroup) Read

```go
func (p *ShapingGroup) Read(iprot thrift.TProtocol) error
```

#### func (*ShapingGroup) ReadField1

```go
func (p *ShapingGroup) ReadField1(iprot thrift.TProtocol) error
```

#### func (*ShapingGroup) ReadField2

```go
func (p *ShapingGroup) ReadField2(iprot thrift.TProtocol) error
```

#### func (*ShapingGroup) ReadField3

```go
func (p *ShapingGroup) ReadField3(iprot thrift.TProtocol) error
```

#### func (*ShapingGroup) String

```go
func (p *ShapingGroup) String() string
```

#### func (*ShapingGroup) Write

```go
func (p *ShapingGroup) Write(oprot thrift.TProtocol) error
```

#### type UnshapeGroupArgs

```go
type UnshapeGroupArgs struct {
	Id    int64  `thrift:"id,1" json:"id"`
	Token string `thrift:"token,2" json:"token"`
}
```


#### func  NewUnshapeGroupArgs

```go
func NewUnshapeGroupArgs() *UnshapeGroupArgs
```

#### func (*UnshapeGroupArgs) GetId

```go
func (p *UnshapeGroupArgs) GetId() int64
```

#### func (*UnshapeGroupArgs) GetToken

```go
func (p *UnshapeGroupArgs) GetToken() string
```

#### func (*UnshapeGroupArgs) Read

```go
func (p *UnshapeGroupArgs) Read(iprot thrift.TProtocol) error
```

#### func (*UnshapeGroupArgs) ReadField1

```go
func (p *UnshapeGroupArgs) ReadField1(iprot thrift.TProtocol) error
```

#### func (*UnshapeGroupArgs) ReadField2

```go
func (p *UnshapeGroupArgs) ReadField2(iprot thrift.TProtocol) error
```

#### func (*UnshapeGroupArgs) String

```go
func (p *UnshapeGroupArgs) String() string
```

#### func (*UnshapeGroupArgs) Write

```go
func (p *UnshapeGroupArgs) Write(oprot thrift.TProtocol) error
```

#### type UnshapeGroupResult

```go
type UnshapeGroupResult struct {
}
```


#### func  NewUnshapeGroupResult

```go
func NewUnshapeGroupResult() *UnshapeGroupResult
```

#### func (*UnshapeGroupResult) Read

```go
func (p *UnshapeGroupResult) Read(iprot thrift.TProtocol) error
```

#### func (*UnshapeGroupResult) String

```go
func (p *UnshapeGroupResult) String() string
```

#### func (*UnshapeGroupResult) Write

```go
func (p *UnshapeGroupResult) Write(oprot thrift.TProtocol) error
```
