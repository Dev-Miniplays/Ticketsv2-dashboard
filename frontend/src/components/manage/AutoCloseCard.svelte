<Card footer="{false}">
  <span slot="title">
    Auto Schließen
  </span>

  <div slot="body" class="body-wrapper">
    <form class="form-wrapper" on:submit|preventDefault={submit}>
      <div class="row do-margin">
        <Checkbox col4={true} label="Aktiviert" bind:value={data.enabled}/>
        <Checkbox col4={true} label="Schließen wenn Benutzer Server verlässt" bind:value={data.on_user_leave}/>
      </div>
      <div class="row" style="justify-content: space-between">
        <div class="col-2" style="flex-direction: row">
          <Duration label="Offen ohne Antwort" badge="Premium" disabled={!isPremium}
                    bind:days={sinceOpenDays} bind:hours={sinceOpenHours} bind:minutes={sinceOpenMinutes}/>
        </div>
        <div class="col-2" style="flex-direction: row">
          <Duration label="Seit Letzter Nachricht" badge="Premium" disabled={!isPremium}
                    bind:days={sinceLastDays} bind:hours={sinceLastHours} bind:minutes={sinceLastMinutes}/>
        </div>
      </div>
      <div class="row">
        <div class="col-1">
          <Button icon="fas fa-paper-plane" fullWidth=true>Speichern</Button>
        </div>
      </div>
    </form>
  </div>
</Card>

<style>
    .form-wrapper {
        display: flex;
        flex-direction: column;
        width: 100%;
        height: 100%;
    }

    .row {
        display: flex;
        width: 100%;
        height: 100%;
    }

    @media only screen and (max-width: 950px) {
        .row {
            flex-direction: column;
        }
    }

    .form-wrapper > .row:not(:last-child) {
        margin-bottom: 1%;
    }
</style>

<script>
    import Card from "../Card.svelte";
    import Checkbox from "../form/Checkbox.svelte";
    import {notifyError, notifySuccess, withLoadingScreen} from "../../js/util";
    import axios from "axios";
    import {API_URL} from "../../js/constants";
    import Duration from "../form/Duration.svelte";
    import {toDays, toHours, toMinutes} from "../../js/timeutil";
    import Button from "../Button.svelte";

    export let guildId;

    let data = {};
    let isPremium = false;

    let sinceOpenDays = 0, sinceOpenHours = 0, sinceOpenMinutes = 0;
    let sinceLastDays = 0, sinceLastHours = 0, sinceLastMinutes = 0;

    async function submit() {
        data.since_open_with_no_response = sinceOpenDays * 86400 + sinceOpenHours * 3600 + sinceOpenMinutes * 60;
        data.since_last_message = sinceLastDays * 86400 + sinceLastHours * 3600 + sinceLastMinutes * 60;

        const res = await axios.post(`${API_URL}/api/${guildId}/autoclose`, data);
        if (res.status !== 200) {
            notifyError(res.data.error);
            return;
        }

        notifySuccess('Auto close settings updated successfully');
    }

    async function loadPremium() {
        const res = await axios.get(`${API_URL}/api/${guildId}/premium`);
        if (res.status !== 200) {
            notifyError(res.data.error);
            return;
        }

        isPremium = res.data.premium;
    }

    async function loadSettings() {
        const res = await axios.get(`${API_URL}/api/${guildId}/autoclose`);
        if (res.status !== 200) {
            notifyError(res.data.error);
            return;
        }

        data = res.data
        update(res.data);
    }

    function update(data) {
        if (data.since_open_with_no_response) {
            sinceOpenDays = toDays(data.since_open_with_no_response);
            sinceOpenHours = toHours(data.since_open_with_no_response);
            sinceOpenMinutes = toMinutes(data.since_open_with_no_response);
        }

        if (data.since_last_message) {
            sinceLastDays = toDays(data.since_last_message);
            sinceLastHours = toHours(data.since_last_message);
            sinceLastMinutes = toMinutes(data.since_last_message);
        }
    }

    withLoadingScreen(async () => await Promise.all([
        loadPremium(),
        loadSettings()
    ]));
</script>
