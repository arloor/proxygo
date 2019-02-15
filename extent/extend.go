package extent

//在windows平台才会有真实的操作
var SetAutoRun func() = func() {}
