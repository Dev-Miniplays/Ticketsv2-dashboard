<div class="parent">
  <div class="content">
    <div class="container">
      <Card footer={false}>
        <span slot="title">Suchst nach Whitelabel Einstellungen?</span>
        <div slot="body" class="body-wrapper">
          <p>Wenn du nach den Whitelabel Einstellungen (Eigener Bot Name und Avatar) suchst, 
             ist dies auf einer anderen Seite, <Navigate to="/whitelabel" styles="link-blue">verfügbar hier</Navigate>.
          </p>
        </div>
      </Card>
    </div>

    <div class="container">
      <Card footer={false}>
        <div slot="title" class="row">
          Farb Schema
          <Badge>Premium</Badge>
        </div>
        <div slot="body" class="body-wrapper">
          <form class="settings-form" on:submit|preventDefault={updateColours}>
            <div class="row colour-picker-row">
              <Colour col3={true} label="Erfolg" bind:value={colours["0"]} disabled={!isPremium} />
              <Colour col3={true} label="Fehlschlag" bind:value={colours["1"]} disabled={!isPremium} />
            </div>

            <div class="row centre">
              <Button icon="fas fa-paper-plane">Speichern</Button>
            </div>
          </form>
        </div>
      </Card>
    </div>
  </div>
</div>

<script>
  import Card from "../components/Card.svelte";
  import {notifyError, notifySuccess, withLoadingScreen} from '../js/util'
  import axios from "axios";
  import {API_URL} from "../js/constants";
  import {setDefaultHeaders} from '../includes/Auth.svelte'
  import {Navigate} from "svelte-router-spa";
  import Colour from "../components/form/Colour.svelte";
  import Button from "../components/Button.svelte";
  import Badge from "../components/Badge.svelte";

  export let currentRoute;
  let guildId = currentRoute.namedParams.id;

  let colours = {};
  let isPremium = false;

  async function updateColours() {
    const res = await axios.post(`${API_URL}/api/${guildId}/customisation/colours`, colours);
    if (res.status !== 200) {
      notifyError(res.data.error);
      return;
    }

    notifySuccess(`Dein Farb Schema wurde gespeichert`);
  }

  async function loadColours() {
    const res = await axios.get(`${API_URL}/api/${guildId}/customisation/colours`);
    if (res.status !== 200) {
      notifyError(res.data.error);
      return;
    }

    colours = res.data;
  }

  async function loadPremium() {
    const res = await axios.get(`${API_URL}/api/${guildId}/premium?include_voting=true`);
    if (res.status !== 200) {
      notifyError(res.data.error);
      return;
    }

    isPremium = res.data.premium;
  }

  withLoadingScreen(async () => {
    setDefaultHeaders(); // TODO: Is this needed?

    await Promise.all([
        loadPremium(),
        loadColours()
    ]);
  });
</script>

<style>
    .parent {
        display: flex;
        justify-content: center;
        width: 100%;
        height: 100%;
    }

    .content {
        display: flex;
        flex-direction: column;
        justify-content: space-between;
        width: 96%;
        height: 100%;
        margin-top: 30px;
    }

    .body-wrapper {
        display: flex;
        flex-direction: column;
        width: 100%;
        height: 100%;
    }

    .row {
        display: flex;
        flex-direction: row;
        width: 100%;
        height: 100%;
        margin-bottom: 2%;
    }

    .col {
        display: flex;
        flex-direction: column;
        width: 100%;
        height: 100%;
        margin-bottom: 2%;
    }

    .container {
        margin-top: 4%;
    }

    .colour-picker-row {
        display: flex;
        flex-direction: row;
        justify-content: space-around;
    }

    .centre {
        justify-content: center;
    }

    @media only screen and (max-width: 950px) {
        .row {
            flex-direction: column;
        }
    }
</style>
