var express = require('express');
var http = require('http');
var path = require('path');
var bodyParser = require('body-parser');

var app = express();
var urlencodedParser = bodyParser.urlencoded({ extended: false })

app.use(express.static(__dirname))
app.set('port',8001);
http.createServer(app).listen(app.get('port'),function () {
    console.log("Express started on http://114.115.235.186:"+app.get('port')+'/WEB-INF/login.html')
});

app.post('/login',urlencodedParser,function (req, res) {
	var userCode = req.body.userCode;
	var pwd = req.body.pwd;
	res.json({"userCode":userCode,"pwd":pwd})
	

});