/**
 * 统计
 * 在网站的没个个网页使用这个js
 * 使用jquery语法
 */

$(document).ready(function () {
    /**
     * 上报用户信息，将访问数据上传打点服务器
     *
     * https://www.cnblogs.com/adolfmc/p/7698364.html
     */

    // gettime(); //js获取当前时间
    // getip(); //js获取客户端ip
    // geturl(); //js获取客户端当前url
    // getrefer(); //js获取客户端当前页面的上级页面的url
    // getuser_agent(); //js获取客户端类型
    // getcookie() //js获取客户端cookie
    // loadXMLDoc();

    // var lock = true; //统计锁，防止页面上报多次，这里去查js锁的使用，示例先不管
    $.get("http://localhost:8000/dig", {
        "time": gettime(),
        "ip": getip(),
        "url": geturl(),
        "refer": getrefer(),
        "ua": getuser_agent(),
    })






})




function gettime(){
    var nowDate = new Date();
    return nowDate.toLocaleString();
}
function geturl(){
    return window.location.href;
}
function getip(){
    return returnCitySN["cip"]+','+returnCitySN["cname"];
}
function getrefer(){
    return document.referrer;
}
function getcookie(){
    return document.cookie;
}
function getuser_agent(){
    return navigator.userAgent;
}
function loadXMLDoc(){
    var xmlhttp;
    if (window.XMLHttpRequest){
        xmlhttp=new XMLHttpRequest();
    }else{
        xmlhttp=new ActiveXObject("Microsoft.XMLHTTP");
    }
    xmlhttp.onreadystatechange=function(){
        if (xmlhttp.readyState==4 && xmlhttp.status==200){
//alert(xmlhttp.responseText);
        }
    } //http://localhost/git_work/log.php //http://localhost:8088/log.php
    xmlhttp.open("POST","http://analysis.wml.com:8088/log.php",true);
    xmlhttp.setRequestHeader("Content-type","application/x-www-form-urlencoded");
    xmlhttp.send("time="+gettime()+"&ip="+getip()+"&url="+geturl()+"&refer="+getrefer()+"&user_agent="+getuser_agent()+"&cookie="+getcookie());
}
