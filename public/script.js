const genBtn = document.getElementById("btn")
const urlInput = document.getElementById("url")
const resP = document.getElementById("result")

genBtn.onclick = (event) => {
  const url = urlInput.value.trim();
  getShortUrl(url)
}

function getShortUrl(baseUrl) {
  obj = {url: baseUrl}
  fetch("/shorten", {
    method: "POST",
    headers: {"Content-Type": "application/json"},
    body: JSON.stringify(obj)
  })
    .then((resp) => resp.json())
    .then((result) => {
        showUrl(result);
    })
    .catch((error) => {
      console.log(error);
    });
}

function showUrl(response) {
  const {
    BaseUrl, ShortUrl
  } = response;
  resP.innerHTML = ShortUrl;
}
