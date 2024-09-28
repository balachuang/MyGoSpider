javascript:(function() {
	var _temp_mask_ = document.createElement("div");
	_temp_mask_.style.backgroundColor = 'rgba(0,0,0,0.5)';
	_temp_mask_.style.pointerEvents = 'none';
	_temp_mask_.style.width = '10000px';
	_temp_mask_.style.height = '10000px';
	_temp_mask_.style.position = 'fixed';
	_temp_mask_.style.left = '-10px';
	_temp_mask_.style.top = '-10px';
	_temp_mask_.style.zIndex = '10000';
	document.getElementsByTagName('body')[0].appendChild(_temp_mask_);
})()