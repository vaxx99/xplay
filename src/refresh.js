var limit="2:00"
if (document.images){
	var parselimit=limit.split(":")
	parselimit=parselimit[0]*60+parselimit[1]*1
}
function beginrefresh(){
	if (!document.images)
		return
	if (parselimit==1)
		window.location.reload()
	else{
		parselimit-=1
		curmin=Math.floor(parselimit/60)
		cursec=parselimit%60
		if (curmin!=0)
			curtime=curmin+" мин. "+cursec+" сек. до обновления!"
		else
			curtime=cursec+" сек. до обновления!"
		window.status=curtime
		setTimeout("beginrefresh()",1000)
	}
}
window.onload=beginrefresh
