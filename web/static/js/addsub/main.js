function getParseFloat(v) {
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
}

$(function () {
    let tg = window.Telegram.WebApp;
    if (window.App.isPrivate) {
        tg.expand();
        tg.MainButton.setText("Добавить подписку");
        Telegram.WebApp.onEvent('mainButtonClicked', function () {
            $form.trigger("submit");
        });
    }

    let $form = $('form'), form = $form[0]
    let $alert = $('[role="alert"]'), alert = $alert[0]
    let $street = $form.find('[name="street_id"]'), street = $street[0]
    let $submit = $form.find('[type="submit"]'), submit = $submit[0]
    let alert_type = '';

    function showAlert(mess, type) {
        $alert.removeClass('alert-' + alert_type);
        $alert.removeClass('d-none');
        $alert.addClass('alert-' + type);
        $alert.html(mess);
        alert_type = type;
        $('html,body').animate({scrollTop: 0}, 500);
    }

    function hideAlert() {
        $alert.addClass('d-none');
        $alert.removeClass('alert-' + alert_type);
    }

    function checkError(code) {
        switch (code) {
            case "INTERNAL":
                showAlert("Internal server error", "danger");
                break;
            case "ALREADY_EXISTS":
                showAlert("Подписка уже существует", "info");
                break;
            case "CHAT_ID_NOT_SET":
                showAlert("Forbidden", "warning");
                break;
        }
    }

    if ($.fn.selectize) {
        $("select").selectize({
            plugins: ["clear_button"],
            persist: true,
            allowEmptyOption: true,
            setFirstOptionActive: false,
            maxOptions: 2500,
        });
    }

    function initButtons() {
        if (window.App.isPrivate) {
            tg.MainButton.show();
        }
        $submit.removeAttr("disabled");
    }

    if (window.App.hasStreet) {
        $.ajax({
            url: window.App.url.streets,
            method: "GET",
            dataType: 'json',
            success: function (data) {
                checkError(data.message);
                if (data.message === "OK") {
                    let s = street.selectize;
                    for (let st of data.result) {
                        s.addOption({value: st.i, text: st.n});
                    }
                    initButtons();
                }
            },
            error: function () {
                showAlert("Не удалось загрузить список улиц", "danger")
            }
        })
    } else {
        initButtons();
    }

    function encodeQueryData(data) {
        const ret = [];
        for (let d in data)
            ret.push(encodeURIComponent(d) + '=' + encodeURIComponent(data[d]));
        return ret.join('&');
    }

    $form.on("submit", function (event) {
        event.preventDefault()
        let m = {};
        let data = new FormData(form);
        for (let entry of data) {
            let n = entry[0]
            let v = entry[1]
            let fv = getParseFloat(v)
            let iv = Math.round(fv);
            if (n.indexOf("house") > -1 && v !== "") {
                m['house'] = v
            } else if (n === "street_id" && iv > 0) {
                m[n] = iv
            } else if (n.indexOf("_from") > -1 && iv > 0) {
                m[n] = iv
            } else if (n.indexOf("_to") > -1 && iv > 0) {
                m[n] = iv
            }
        }

        hideAlert()
        let urlParams = encodeQueryData(m);
        if (urlParams !== "") {
            $.ajax({
                url: window.App.url.sub_add + "?" + window.App.url.query,
                method: "POST",
                dataType: 'json',
                data: JSON.stringify(m),
                success: function (data) {
                    checkError(data.message);
                    if (data.message === "OK") {
                        showAlert("Подписка успешно добавлена", "success")
                    }
                },
                error: function () {
                    showAlert("Не удалось добавить подписку", "danger")
                }
            })
        } else {
            showAlert("Форма не заполнена", "warning");
        }
    });

    if (typeof window.App.slider !== "undefined") {
        for (let params of window.App.slider) {
            new Slider(params);
        }
    }
});
