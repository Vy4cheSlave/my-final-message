function request() {
    controller.abort();
    controller = new AbortController;

    const recievedFile = document.getElementById('receivedFile');
    const langSelect = document.getElementById('langSelect');

    var data = new FormData()
    data.append('receivedFile', recievedFile.files[0]);
    data.append('lang', langSelect.value);

    alert('Пожалуйста ожидайте ответа!')

    fetch('/serve', {
        method: 'POST',
        body: data,
        signal: controller.signal,
    }).then((response) => {
        console.log(response);
        return response.json();
    })
        .then((result) => {
            var requestResult = '<div class="internal-object">Название файла <input id="title" type="text" name="title" class="buttons"></div><div class="internal-object"><textarea id="textarea" name="body" class="textarea"></textarea></div><button id="button" class="buttons">Save</button>';
            document.getElementById("requestResult").innerHTML = requestResult;

            function onScroll(evt) {
                prevScrollPos = this.scrollTop;
            }

            var title = document.getElementById('title');
            title = result['title'];

            var body = document.getElementById('textarea');
            body.value = result['body'];
            body.addEventListener('scroll', onScroll);

            $('.textarea').highlightWithinTextarea({
                highlight: ''
            });

            alert('Сервер обработал ваш запрос!')

            var player = document.getElementById('audioplayer');

            var wordsPointers = result['wordspointers'];
            var sliceWords = result['slicewords'];
            var lengthArr = wordsPointers.length;
            function print(event) {
                var timepoint = wordsPointers.findIndex((element) => element > player.currentTime);
                if (timepoint > 0) {
                    timepoint = (timepoint - 1);
                } else if (timepoint < 0) {
                    timepoint = (lengthArr - 1);
                }
                // player.currentTime
                $('.textarea').highlightWithinTextarea({
                    highlight: sliceWords[timepoint]
                });

                body.scrollTo(0, prevScrollPos);
            }
            player.addEventListener("timeupdate", print);

            var button = document.getElementById('button')
            button.addEventListener('click', function (e) {
                var textareaVal = document.getElementById('textarea').value
                var filename = document.getElementById('title').value
                download(textareaVal, filename)
            })

            function download(textareaVal, filename) {
                var element = document.createElement('a')
                element.setAttribute('href', 'data:text/plain;charset=utf-8,' + encodeURIComponent(textareaVal));
                element.setAttribute('download', filename);
                element.style.display = 'none';
                document.body.appendChild(element)
                element.click()
                document.body.removeChild(element)
            }
        }).catch((err) => {
            alert(err);
            document.getElementById("requestResult").innerHTML = '';
        });
}