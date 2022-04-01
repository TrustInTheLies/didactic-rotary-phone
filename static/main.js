let body = document.querySelector("body");
let btn = document.querySelector(".btn");
let songs = [];
const api_url = "http://localhost:8080";

const getToken = async () => {
  let request = await fetch(`${api_url}/request-page`);
  return await request.json();
};

const getData = async (offset = 0) => {
  let request = await fetch(
    `https://api.spotify.com/v1/me/tracks?limit=50&offset=${offset}`,
    {
      headers: {
        Authorization: `Bearer ${await getToken()}`,
        "Content-Type": "application/json",
      },
    }
  );
  return await request.json();
};

const createArr = async () => {
  let { total, limit, offset } = await getData();
  const pages = total / limit;
  for (let i = 0; i <= pages; i++) {
    const { items } = await getData(offset);
    songs.push(...items);
    offset += 50;
  }
  return songs;
};

const createList = () => {
  let list = [];
  songs.forEach((song) => {
    let item = {
      title: song.track.name,
      artist: song.track.artists[0].name,
    };
    list.push(item);
  });
  return list;
};

const sendList = async () => {
  await fetch(`${api_url}/send-list`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(createList()),
  });
};

const getFile = async () => {
  let request = await fetch(`${api_url}/get-file`);
  let response = request.blob();
  let url = URL.createObjectURL(await response);
  let link = document.createElement("a");
  link.href = url;
  link.setAttribute("download", "songs.csv");
  link.textContent = "Загрузить";
  body.append(link);
  btn.setAttribute("disabled", true);
};

btn.addEventListener("click", async () => {
  await createArr();
  await sendList();
  await getFile();
});
