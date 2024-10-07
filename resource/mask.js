javascript:(function() {
	var __tl__ = document.getElementById('__temp_light__');
	if (__tl__) {
		__tl__.remove();
	}else{
		var __tl__ = document.createElement('div');
		__tl__.id = '__temp_light__';
		__tl__.style.backgroundColor = 'rgba(0,0,0,0.5)';
		__tl__.style.pointerEvents = 'none';
		__tl__.style.width = '10000px';
		__tl__.style.height = '10000px';
		__tl__.style.position = 'fixed';
		__tl__.style.left = '-10px';
		__tl__.style.top = '-10px';
		__tl__.style.zIndex = '10000';
		document.getElementsByTagName('body')[0].appendChild(__tl__);
	}
})()