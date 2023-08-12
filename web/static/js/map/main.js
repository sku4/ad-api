function showPreloader() {
    $("body").addClass('show-preloader');
}

function hidePreloader() {
    $("body").removeClass('show-preloader');
}

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

function declOfNum(n, text_forms) {
    n = Math.abs(n) % 100;
    let n1 = n % 10;
    if (n > 10 && n < 20) {
        return text_forms[2];
    }
    if (n1 > 1 && n1 < 5) {
        return text_forms[1];
    }
    if (n1 === 1) {
        return text_forms[0];
    }
    return text_forms[2];
}

$(function () {
    let $body = $("body");
    let $burgerFilter = $(".burger-filter");
    let $contentFilter = $(".content .filter");
    let $burgerMenuCont = $('[for="burger-menu-checkbox"]');
    let $burgerMenu = $("#burger-menu-checkbox");
    let $contentMenu = $(".content-menu");
    let showFilterClass = "show-filter"
    let showMenuClass = "show-menu"
    let $form = $('form[name="filter"]'), form = $form[0]
    let $alert = $('[role="alert"]'), alert = $alert[0]
    let $street = $form.find('[name="street_id"]'), street = $street[0]
    let $formTexts = $form.find('[type="text"]')
        .not('.slider-from')
        .not('.slider-to')
    let $reset = $form.find('.reset'), resetBtn = $reset[0]
    let $map = $('#map');
    let eventStreetLoad = new Event("onStreetLoaded");
    let eventFilterLoad = new Event("onFilterLoaded");
    let wnumb = wNumb(window.App.wnumb);
    let popup = null;

    function showAlert(mess, type) {
        hidePreloader()
        $alert.html(mess);
    }

    function hideAlert() {
        hidePreloader()
        $alert.html("");
    }

    function checkError(code) {
        switch (code) {
            case "INTERNAL":
                showAlert("Internal server error", "danger");
                break;
        }
    }

    // init map progress-bar
    let progress = document.getElementById('progress');
    let progressBar = document.getElementById('progress-bar');

    function updateProgressBar(processed, total, elapsed, layersArray) {
        if (elapsed > 1000) {
            // if it takes more than a second to load, display the progress bar:
            progress.style.display = 'block';
            progressBar.style.width = Math.round(processed / total * 100) + '%';
        }

        if (processed === total) {
            // all markers processed - hide the progress bar:
            progress.style.display = 'none';
        }
    }

    // init map
    let map = L.map('map').setView([53.9006, 27.5590], 11);
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        maxZoom: 19,
        minZoom: 9,
        attribution: '&copy; <a href="http://www.openstreetmap.org/copyright">OpenStreetMap</a>',
        detectRetina: true,
    }).addTo(map);
    let mapMarkers = L.markerClusterGroup({
        spiderfyOnMaxZoom: false,
        chunkedLoading: true,
        zoomToBoundsOnClick: false,
        chunkInterval: 200,
        chunkDelay: 50,
        chunkProgress: updateProgressBar,
    });
    map.addLayer(mapMarkers);

    mapMarkers.on('click', function (a) {
        drawPopup([a.layer], a.layer.getLatLng(), false);
    });

    mapMarkers.on('clusterclick', function (a) {
        if (a.layer.getChildCount() < 10 || map.getZoom() === map.getMaxZoom()) {
            drawPopup(a.layer.getAllChildMarkers(), a.layer.getLatLng(), true);
        } else {
            a.layer.zoomToBounds();
        }
    });

    map.on('zoomstart', function () {
        if (popup !== null) {
            if (typeof popup.options.isCluster !== 'undefined') {
                if (popup.options.isCluster) {
                    map.closePopup();
                }
            }
        }
    });

    function loadMap(locs) {
        let markerList = [];
        let count = 0;
        let max_links = {};
        for (let i = 0; i < locs.length; i++) {
            let l = locs[i];
            let ids = [];
            if (typeof l['s'] !== 'undefined') {
                ids = l['s'];
            } else {
                ids = [l['i']];
            }
            count += ids.length;
            if (typeof max_links[ids.length] === 'undefined') {
                max_links[ids.length] = 0
            }
            max_links[ids.length]++;
            let marker = L.marker(L.latLng(l['t'], l['g']), {ids: ids});
            markerList.push(marker);
        }
        // console.log(max_links);
        /*if (count > 0) {
            console.log('found ' + ((count - markerList.length) * 100 / count).toFixed(0) + "% (" +
                ((count - markerList.length) / 2) + ") duplicates");
        }*/
        map.closePopup();
        mapMarkers.clearLayers();
        mapMarkers.addLayers(markerList);
    }

    function drawPopup(markers, position, isCluster) {
        let groups = []
        for (let i in markers) {
            if (!markers.hasOwnProperty(i))
                continue;
            let marker = markers[i];
            groups.push({ids: marker.options.ids});
        }

        let offset = L.point(0, -7);
        let maxHeight = 285;
        if (isCluster === false) {
            offset = L.point(0, -31);
            maxHeight = 255;
        }
        let content = '<div class="ad-popup d-flex align-items-center justify-content-center">' +
            '<div class="lds-ripple"><div></div><div></div></div>' +
            '</div>';

        popup = L.popup([position.lat, position.lng], {
            maxWidth: 250,
            minWidth: 250,
            closeButton: false,
            maxHeight: maxHeight,
            offset: offset,
            content: content,
            isCluster: isCluster,
        });

        popup.openOn(map);
        loadAds(groups, popup, isCluster);
    }

    let adsXhr = null;

    function loadAds(groups, popup, isCluster) {
        let adsWord = 'объявления';
        if (isCluster === false) {
            adsWord = 'объявление';
        }
        let notLoadContent = '<div class="ad-popup d-flex align-items-center justify-content-center">' +
            '<p>Не удалось загрузить ' + adsWord + '</p>' +
            '</div>';

        if (adsXhr !== null) {
            adsXhr.abort();
        }
        adsXhr = $.ajax({
            url: window.App.url.ads,
            method: "POST",
            dataType: 'json',
            data: JSON.stringify({
                groups: groups
            }),
            success: function (data) {
                if (data.message === "OK") {
                    popup.setContent(getMarkersHtml(data.result, isCluster));
                } else {
                    popup.setContent(notLoadContent);
                }
            },
            error: function () {
                popup.setContent(notLoadContent);
            }
        });
    }

    function getMarkersHtml(ads, isCluster) {
        let some = 'some';
        if (isCluster === false) {
            some = '';
        }

        let content = '<div class="ad-popup ' + some + ' d-flex flex-column justify-content-between">';
        let adsContent = '';
        for (let i = 0; i < ads.length; i++) {
            let ad = ads[i];
            let isLast = (i === ads.length - 1)
            let adContent = '<div class="ad">';
            let img = '';
            if (typeof ad.photo !== 'undefined' && ad.photo !== "") {
                img = 'style="background-image: url(' + ad.photo + ')"';
            }
            // 54 022 $   —   1 190 $ / м²
            let price = '';
            if (typeof ad.price !== 'undefined' && ad.price > 0) {
                price = '<span>' + wnumb.to(ad.price) + '</span>';
            }
            if (typeof ad.price_m2 !== 'undefined' && ad.price_m2 > 0) {
                if (price !== "") {
                    price += "&nbsp;&nbsp;—&nbsp;&nbsp;" + wnumb.to(ad.price_m2) + ' / м²';
                } else {
                    price += this.wnumb.to(ad.price_m2) + ' / м²';
                }
            }
            if (price !== "") {
                price = '<div class="d-flex align-items-end p-2 price">' + price + '</div>';
            }
            let a_url = ad.urls[0].u
            let img_class = ''
            let links = ''
            if (ad.urls.length > 1) {
                ad.urls = ad.urls.slice(0, 12);
                a_url = '#';
                img_class = 'multi_links';
                let links_icons = '';
                let count_on_row = 4;
                let links_m = 'm-2';
                let links_more = '';
                if (ad.urls.length <= 3) {
                    links_m = 'm-3';
                    links_more = 'more';
                }
                for (let j = 0; j < ad.urls.length; j++) {
                    let links_url = ad.urls[j];
                    if (j > 0 && j % count_on_row === 0) {
                        links_icons += '</div><div class="flex-row">';
                    }
                    links_icons += '<a class="' + links_more + ' ' + links_m + ' ' + links_url.p +
                        '" href="' + links_url.u + '" target="_blank"></a>';
                }
                links = '<div class="d-flex flex-column align-items-center justify-content-center d-none p-2 links">' +
                    '<div class="flex-row">' +
                    links_icons +
                    '</div>' +
                    '</div>';
            }
            let address = '';
            if (typeof ad.address !== 'undefined' && ad.address !== "") {
                address = ad.address;
            }
            if (typeof ad.year !== 'undefined' && ad.year > 0) {
                if (address !== "") {
                    address += '&nbsp;&nbsp;<span>(' + ad.year + ' г.п.)</span>';
                } else {
                    address += '<span>' + ad.year + ' г.п.</span>';
                }
            }
            if (address !== "") {
                address = '<div class="row-address mx-2 mb-1">' + address + '</div>';
            }
            // 2 комн   45.3 / 22.6 / 10 м²   этаж 15/16
            let m2 = '';
            if (typeof ad.rooms !== 'undefined' && ad.rooms > 0) {
                m2 += '<div>' + ad.rooms + ' комн</div>';
            }
            let m2sr = [];
            if (typeof ad.m2_main !== 'undefined' && ad.m2_main > 0) {
                m2sr.push(ad.m2_main);
            }
            if (typeof ad.m2_living !== 'undefined' && ad.m2_living > 0) {
                m2sr.push(ad.m2_living);
            }
            if (typeof ad.m2_kitchen !== 'undefined' && ad.m2_kitchen > 0) {
                m2sr.push(ad.m2_kitchen);
            }
            if (m2sr.length > 0) {
                m2 += '<div>' + m2sr.join(' / ') + ' м²</div>';
            }
            let floor = [];
            if (typeof ad.floor !== 'undefined' && ad.floor > 0) {
                floor.push(ad.floor);
            }
            if (typeof ad.floors !== 'undefined' && ad.floors > 0) {
                floor.push(ad.floors);
            }
            if (floor.length > 0) {
                m2 += '<div>' + floor.join('/') + ' эт</div>';
            }
            if (m2 !== "") {
                m2 = '<div class="row-m2 d-flex justify-content-between mx-2 mb-1">' + m2 + '</div>';
            }
            let bathroom = '';
            if (typeof ad.bathroom !== 'undefined' && ad.bathroom !== "") {
                bathroom = "с/у " + ad.bathroom.toLowerCase();
            }
            if (bathroom !== "") {
                bathroom = '<div class="row-bath d-flex mx-2 mb-1">' + bathroom + '</div>';
            }
            adContent += '<div class="col-cont d-flex flex-column">' +
                '<div class="row-img mb-2 ' + img_class + '">' +
                (ad.urls.length === 1 ? '<a href="' + a_url + '" target="_blank">' : '') +
                '       <div class="d-flex h-100 img" ' + img + '></div>' +
                '       ' + price +
                '       ' + links +
                (ad.urls.length === 1 ? '</a>' : '') +
                '</div>' +
                address +
                m2 +
                bathroom +
                '</div>';

            adContent += '</div>';
            adsContent += adContent;
        }

        if (isCluster === false) {
            content += adsContent;
        } else {
            content += '<div class="ads">' + adsContent + '</div>';
            content += '<div class="count d-flex justify-content-center p-2">' +
                ads.length.toLocaleString('ru-RU', {minimumFractionDigits: 0}) +
                " " + declOfNum(ads.length, ['объявление', 'объявления', 'объявлений'])
                + '</div>';
        }
        content += '</div>';

        return content;
    }

    // init selectize
    if ($.fn.selectize) {
        $("select").selectize({
            plugins: ["clear_button"],
            persist: true,
            allowEmptyOption: true,
            setFirstOptionActive: false,
            maxOptions: 2500,
        });
    }
    let $clearSelect = $form.find('.selectize-control.plugin-clear_button .clear')

    if (window.App.hasStreet) {
        showPreloader();
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
                    if (window.App.valStreet !== null) {
                        s.setValue(window.App.valStreet, true)
                        $clearSelect.show();
                    }
                }
                hideAlert();
                document.dispatchEvent(eventStreetLoad);
            },
            error: function () {
                showAlert("Не удалось загрузить список улиц", "danger")
            }
        })
    } else {
        document.dispatchEvent(eventStreetLoad);
    }

    let locsXhr = null;

    function filterLocations() {
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

        if (locsXhr !== null) {
            locsXhr.abort();
        }

        showPreloader();
        locsXhr = $.ajax({
            url: window.App.url.locations + "?" + window.App.url.query,
            method: "POST",
            dataType: 'json',
            data: JSON.stringify(m),
            success: function (data) {
                checkError(data.message);
                if (data.message === "OK") {
                    showAlert(data.result.length.toLocaleString('ru-RU', {minimumFractionDigits: 0}) +
                        " " + declOfNum(data.result.length, ['объявление', 'объявления', 'объявлений']),
                        "success")
                    loadMap(data.result);
                }
            },
            error: function () {
                showAlert("Не удалось получить объявления", "danger");
            }
        });
    }

    $form.on("submit", function (event) {
        event.preventDefault()
    });

    $form.find("select").on("change", function () {
        filterLocations();
    });

    $formTexts.on("change", function () {
        filterLocations();
    });

    document.addEventListener('onSliderChange', function (event) {
        filterLocations();
    });

    let sliders = []
    if (typeof window.App.slider !== "undefined") {
        for (let params of window.App.slider) {
            let s = new Slider(params);
            sliders.push(s)
        }
    }

    document.addEventListener('onStreetLoaded', function (event) {
        document.dispatchEvent(eventFilterLoad);
    });

    document.addEventListener('onFilterLoaded', function (event) {
        filterLocations();
    });

    function reset() {
        if (window.App.hasStreet) {
            street.selectize.setValue("", true)
            $clearSelect.hide();
        }
        for (const s of sliders) {
            s.reset()
        }
        $formTexts.each(function (index) {
            $(this)[0].value = "";
        });
    }

    $reset.on("click", function (event) {
        event.preventDefault()
        reset();
        filterLocations();
    });

    $burgerMenu.on("change", function (e) {
        if ($body.hasClass(showMenuClass)) {
            $body.removeClass(showMenuClass);
        } else {
            $body.addClass(showMenuClass);
        }
        $contentMenu.slideToggle("fast");
    });

    $burgerFilter.on("click", function (e) {
        e.preventDefault();
        if ($body.hasClass(showFilterClass)) {
            $body.removeClass(showFilterClass);
            $burgerFilter.removeClass('rotate-center');
            $burgerFilter.addClass('rotate-center-reverse');
        } else {
            $body.addClass(showFilterClass);
            $burgerFilter.removeClass('rotate-center-reverse');
            $burgerFilter.addClass('rotate-center');
        }
    });

    // click outside
    $(document).mouseup(function (e) {
        if (!$contentFilter.is(e.target) && $contentFilter.has(e.target).length === 0 &&
            !$burgerFilter.is(e.target) && $burgerFilter.has(e.target).length === 0) {
            let hasClass = $body.hasClass(showFilterClass);
            $body.removeClass(showFilterClass);
            $burgerFilter.removeClass('rotate-center');
            if (hasClass) {
                $burgerFilter.addClass('rotate-center-reverse');
            }
        }
        if (!$contentMenu.is(e.target) && $contentMenu.has(e.target).length === 0 &&
            !$burgerMenuCont.is(e.target) && $burgerMenuCont.has(e.target).length === 0) {
            $contentMenu.hide();
            $body.removeClass(showMenuClass);
            $burgerMenu.prop("checked", false);
        }
    });

    $(window).resize(function () {
        let $this = $(this),
            w = $this.width();
        if (w > 768) {
            $body.removeClass(showFilterClass);
            $burgerFilter.removeClass('rotate-center');
            $burgerFilter.removeClass('rotate-center-reverse');
            $contentMenu.hide();
            $body.removeClass(showMenuClass);
            $burgerMenu.prop("checked", false);
        }
    });

    // click multi links
    $map.delegate(".ad-popup .ad .row-img.multi_links", "click", function (event) {
        $map.find('.ad-popup .ad .row-img .links').addClass('d-none');
        $map.find('.ad-popup .ad .row-img.multi_links').removeClass('opened');
        $(this).find('.links').removeClass('d-none');
        $(this).addClass('opened');
    });

    // click outside multi links
    $(document).mouseup(function (e) {
        $cont = $map.find('.ad-popup .ad .row-img.multi_links');
        if (!$cont.is(e.target) && $cont.has(e.target).length === 0) {
            $map.find('.ad-popup .ad .row-img .links').addClass('d-none');
            $map.find('.ad-popup .ad .row-img.multi_links').removeClass('opened');
        }
    });
});
