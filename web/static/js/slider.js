(function (window) {
    'use strict';

    if (window.Slider)
        return;

    window.Slider = function (arParams) {
        this.obSlider = null;
        this.obSliderFrom = null;
        this.obSliderTo = null;
        this.obSliderConnect = null;
        this.min = 0;
        this.max = 0;
        this.max_percent_step = 0;
        this.wnumb = null;
        this.eventChange = null;
        this.initSlider = false;
        this.resetSlider = false;
        if (typeof arParams === 'object') {
            this.params = arParams;
        }
        this.init()
    };

    window.Slider.prototype = {
        initObjects: function () {
            this.obSlider = document.getElementById('slider-' + this.params.code);
            this.obSliderFrom = document.getElementById('slider-' + this.params.code + '-from');
            this.obSliderTo = document.getElementById('slider-' + this.params.code + '-to');
        },

        init: function () {
            this.initObjects();
            this.eventChange = new Event("onSliderChange");

            let max_percent = 0;
            this.max_percent_step = 1;
            for (let code in this.params.range) {
                if (!this.params.range.hasOwnProperty(code))
                    continue;
                let range = this.params.range[code];
                if (code === "min") {
                    this.min = range[0];
                } else if (code === "max") {
                    this.max = range[0];
                } else {
                    let percent = this.getParseFloat(code);
                    if (max_percent < percent && range.length > 1) {
                        max_percent = percent;
                        this.max_percent_step = range[1];
                    }
                }
            }
            this.max += this.max_percent_step;
            this.params.range['max'][0] = this.max;

            this.wnumb = wNumb(this.params.wnumb)
            let start_from = this.min
            let start_to = this.max
            if (typeof this.params.from.start !== 'undefined') {
                start_from = this.params.from.start
            }
            if (typeof this.params.to.start !== 'undefined') {
                start_to = this.params.to.start
            }
            noUiSlider.create(this.obSlider, {
                start: [start_from, start_to],
                connect: true,
                range: this.params.range,
                format: this.wnumb,
            });

            this.obSliderConnect = $(this.obSlider).find(".noUi-connect")[0];
            this.obSlider.noUiSlider.on('update', $.proxy(this.onUpdate, this));
            this.obSlider.noUiSlider.on('set', $.proxy(this.onChange, this));
            this.obSliderFrom.addEventListener('change', $.proxy(this.onFromChange, this));
            this.obSliderTo.addEventListener('change', $.proxy(this.onToChange, this));
            this.initSlider = true;
        },

        onFromChange: function () {
            let v = this.obSliderFrom.value;
            let res = this.checkRange(v)
            if (res) {
                this.obSlider.noUiSlider.set([v, this.max]);
            } else {
                this.obSlider.noUiSlider.set([v, null]);
            }
        },

        onToChange: function () {
            let v = this.obSliderTo.value;
            this.checkRange(v)
            this.obSlider.noUiSlider.set([null, v]);
        },

        getParseFloat: function (v) {
            if (typeof v === 'string') {
                v = v.replace(new RegExp(",", 'g'), '.');
                v = v.replace(/[^\d.]+/g, '');
            }
            let val = parseFloat(v);
            if (val === null || typeof val === 'undefined' || isNaN(val)) {
                val = 0;
            }
            val = parseFloat(val.toFixed(2));

            return val;
        },

        checkRange: function (val) {
            let i = Math.round(this.getParseFloat(val));
            if (i > this.max) {
                this.max = i + this.max_percent_step;
                this.params.range['max'][0] = this.max;
                this.obSliderTo.setAttribute("placeholder", this.wnumb.to(i));
                this.obSlider.noUiSlider.updateOptions({
                        range: this.params.range,
                    }, false
                );
                return true
            }
            return false
        },

        onUpdate: function (values, handle) {
            let formatValues = [this.obSliderFrom, this.obSliderTo];
            let i = this.getParseFloat(values[handle])
            let s = values[handle]
            switch (handle) {
                case 0:
                    if (i === this.min) {
                        s = ""
                    } else if (i === this.max) {
                        s = this.wnumb.to(this.max - this.max_percent_step)
                    }
                    break;
                case 1:
                    if (i === this.max) {
                        s = ""
                    }
                    break;
            }
            formatValues[handle].value = s;

            let from = this.getParseFloat(values[0]);
            let to = this.getParseFloat(values[1]);
            if (from > this.min || to < this.max) {
                this.obSliderConnect.classList.add("active");
            } else {
                this.obSliderConnect.classList.remove("active");
            }
        },

        onChange: function () {
            if (this.initSlider && !this.resetSlider) {
                document.dispatchEvent(this.eventChange);
            }
        },

        reset: function () {
            this.resetSlider = true
            this.obSlider.noUiSlider.set([this.min, this.max]);
            this.resetSlider = false
        },
    };
})(window);
