<div class="content">
    <Card footer footerRight>
        <span slot="title">Formulare</span>
        <div slot="body" class="body-wrapper">
            <div class="section">
                <h2 class="section-title">Neues Formular erstellen</h2>

                <form on:submit|preventDefault={createForm}>
                    <div class="row" id="creation-row">
                        <Input placeholder="Formular Titel" col3={true} bind:value={newTitle}/>
                        <div id="create-button-wrapper">
                            <Button icon="fas fa-paper-plane" fullWidth={windowWidth <= 950}>Erstellen</Button>
                        </div>
                    </div>
                </form>
            </div>
            <div class="section">
                <h2 class="section-title">Formulare verwalten</h2>

                {#if editingTitle && activeFormId !== null}
                    <div class="row form-name-edit-wrapper">
                        <Input col4 label="Form Title" placeholder="Formular Titel" bind:value={renamedTitle}/>
                        <div class="form-name-save-wrapper">
                            <Button icon="fas fa-floppy-disk" fullWidth={windowWidth <= 950} on:click={updateTitle}>
                                Speichern
                            </Button>
                        </div>
                    </div>
                {:else}
                    <div class="row form-select-row">
                        <div class="multiselect-super">
                            <Dropdown col1 bind:value={activeFormId}>
                                <option value={null}>Formular auswählen...</option>
                                {#each forms as form}
                                    <option value="{form.form_id}">{form.title}</option>
                                {/each}
                            </Dropdown>
                        </div>

                        {#if activeFormId !== null}
                            <Button on:click={() => editingTitle = true}>Formular umbenennen</Button>
                            <Button danger type="button"
                                    on:click={() => deleteForm(activeFormId)}>Formular {activeFormTitle} löschen</Button>
                        {/if}
                    </div>
                {/if}

                <div class="manage">
                    {#if activeFormId !== null}
                        {#each forms.find(form => form.form_id === activeFormId).inputs as input, i (input)}
                            <div animate:flip="{{duration: 500}}">
                                <FormInputRow data={input} formId={activeFormId}
                                              withSaveButton={true} withDeleteButton={true} withDirectionButtons={true}
                                              index={i} {formLength}
                                              on:delete={() => deleteInput(activeFormId, input)}
                                              on:move={(e) => changePosition(activeFormId, input, e.detail.direction)}/>
                            </div>
                        {/each}
                    {/if}

                    {#if activeFormId !== null}
                        <div class="row"
                             style="justify-content: center; align-items: center; gap: 10px; margin-top: 10px">
                            <hr class="fill">
                            <div class="row add-input-container" class:add-input-disabled={formLength >= 5}>
                                <i class="fas fa-plus"></i>
                                <a on:click={addInput}>Neues Feld</a>
                            </div>
                            <hr class="fill">
                        </div>
                    {/if}
                </div>
            </div>
        </div>

        <div slot="footer">
            <Button type="submit" icon="fas fa-floppy-disk" disabled={formLength === 0} on:click={saveInputs}>
                Speichern
            </Button>
        </div>
    </Card>
</div>

<svelte:window bind:innerWidth={windowWidth}/>

<script>
    import Card from "../components/Card.svelte";
    import {notifyError, notifySuccess, nullIfBlank, withLoadingScreen} from '../js/util'
    import Button from "../components/Button.svelte";
    import axios from "axios";
    import {API_URL} from "../js/constants";
    import {setDefaultHeaders} from '../includes/Auth.svelte'
    import Input from "../components/form/Input.svelte";
    import Dropdown from "../components/form/Dropdown.svelte";
    import FormInputRow from "../components/manage/FormInputRow.svelte";
    import {flip} from "svelte/animate";

    export let currentRoute;
    let guildId = currentRoute.namedParams.id;

    let defaultTeam = {id: 'default', name: 'Default'};

    let newTitle;
    let forms = [];
    let toDelete = {};
    let activeFormId = null;
    $: activeFormTitle = activeFormId !== null ? forms.find(f => f.form_id === activeFormId).title : 'Unknown';

    $: formLength = activeFormId !== null ? forms.find(f => f.form_id === activeFormId).inputs.length : 0;

    let editingTitle = false;
    let renamedTitle = "";
    $: activeFormId, reflectTitle();

    function reflectTitle() {
        renamedTitle = activeFormId !== null ? forms.find(f => f.form_id === activeFormId).title : null;
    }

    $: windowWidth = 0;

    function getForm(formId) {
        return forms.find(form => form.form_id === formId);
    }

    async function updateTitle() {
        const res = await axios.patch(`${API_URL}/api/${guildId}/forms/${activeFormId}`, {title: renamedTitle});
        if (res.status !== 200) {
            notifyError('Fehler beim Aktualisieren des Formulartitels');
            return;
        }

        editingTitle = false;
        getForm(activeFormId).title = renamedTitle;
        forms = forms;

        notifySuccess('Formular Titel Aktualisiert');
    }

    async function createForm() {
        let data = {
            title: newTitle,
        };

        const res = await axios.post(`${API_URL}/api/${guildId}/forms`, data);
        if (res.status !== 200) {
            notifyError(res.data.error);
            return;
        }

        notifySuccess(`Formular ${newTitle} wurde erstellt`);
        newTitle = '';

        let form = res.data;
        form.inputs = [];

        activeFormId = null; // Error thrown from {#each forms.find} if we don't temporarily set this to null?
        forms = [...forms, form];
        activeFormId = form.form_id;

        addInput();
    }

    async function deleteForm(id) {
        const res = await axios.delete(`${API_URL}/api/${guildId}/forms/${id}`);
        if (res.status !== 200) {
            notifyError(res.data.error);
            return;
        }

        notifySuccess(`Formular erfolgreich gelöscht`);

        forms = forms.filter(form => form.form_id !== id);
        if (forms.length > 0) {
            activeFormId = forms[0].form_id;
        } else {
            activeFormId = null;
        }
    }

    function addInput() {
        const form = getForm(activeFormId);
        if (form.inputs.length >= 5) return;

        const input = {
            form_id: activeFormId,
            position: form.inputs.length + 1,
            style: "1",
            label: "",
            placeholder: "",
            required: true,
            min_length: 0,
            max_length: 1024,
            is_new: true,
        };

        form.inputs = [...form.inputs, input];
        forms = forms;
    }

    async function deleteInput(formId, input) {
        let form = getForm(formId);

        let idx = form.inputs.findIndex((i) => i === input);
        form.inputs.splice(idx, 1);
        for (let i = idx; i < form.inputs.length; i++) {
            form.inputs[i].position--;
        }

        forms = forms;

        if (!input.is_new) {
            if (toDelete[formId] === undefined) {
                toDelete[formId] = [];
            }

            toDelete[formId] = [...toDelete[formId], input.id];
        }
    }

    function changePosition(formId, input, direction) {
        const form = getForm(formId);
        let idx = form.inputs.findIndex((i) => i === input);

        let inputs = form.inputs;
        if (direction === "up") {
            [inputs[idx - 1].position, inputs[idx].position] = [inputs[idx].position, inputs[idx - 1].position];
            [inputs[idx - 1], inputs[idx]] = [inputs[idx], inputs[idx - 1]];
        } else if (direction === "down") {
            [inputs[idx + 1].position, inputs[idx].position] = [form.inputs[idx].position, form.inputs[idx + 1].position];
            [inputs[idx + 1], inputs[idx]] = [form.inputs[idx], form.inputs[idx + 1]];
        }

        forms = forms;
    }

    async function saveInputs() {
        const form = getForm(activeFormId);

        const data = {
            "create": form.inputs.filter(i => i.is_new === true)
                .map(i => ({...i, style: parseInt(i.style), placeholder: nullIfBlank(i.placeholder)})),
            "update": form.inputs.filter(i => !i.is_new)
                .map(i => ({...i, style: parseInt(i.style), placeholder: nullIfBlank(i.placeholder)})),
            "delete": toDelete[activeFormId] || [],
        }

        const res = await axios.patch(`${API_URL}/api/${guildId}/forms/${activeFormId}/inputs`, data);
        if (res.status !== 204) {
            notifyError(res.data.error);
            return;
        }

        toDelete = {};

        const formId = activeFormId;
        await loadForms();
        activeFormId = formId;

        notifySuccess('Formular erfolgreich aktualisiert');
    }

    async function loadForms() {
        const res = await axios.get(`${API_URL}/api/${guildId}/forms`);
        if (res.status !== 200) {
            notifyError(res.data.error);
            return;
        }

        forms = res.data || [];
        forms.flatMap(f => f.inputs).forEach(i => {
            i.style = i.style.toString();
        });

        if (forms.length > 0) {
            activeFormId = forms[0].form_id;
        }
    }

    withLoadingScreen(async () => {
        setDefaultHeaders();
        await loadForms();
    });
</script>

<style>
    .content {
        display: flex;
        width: 100%;
        height: 100%;
    }

    .body-wrapper {
        display: flex;
        flex-direction: column;
        width: 100%;
        height: 100%;
        padding: 1%;
    }

    .section {
        display: flex;
        flex-direction: column;
        width: 100%;
        height: 100%;
    }

    .section:not(:first-child) {
        margin-top: 2%;
    }

    .section-title {
        font-size: 36px;
        font-weight: bolder !important;
    }

    h3 {
        font-size: 28px;
        margin-bottom: 4px;
    }

    hr.fill {
        border-top: 1px solid #777;
        border-bottom: 0;
        border-left: 0;
        border-right: 0;
        flex: 1;
    }

    .row {
        display: flex;
        flex-direction: row;
        justify-content: space-between;
        width: 100%;
        height: 100%;
    }

    .form-select-row {
        justify-content: flex-start;
        gap: 12px;
        max-height: 40px;
    }

    .form-name-edit-wrapper {
        justify-content: flex-start;
        gap: 12px;
    }

    .form-name-save-wrapper {
        height: 48px;
        align-self: flex-end;
    }

    .multiselect-super {
        width: 31%;
    }

    .manage {
        display: flex;
        flex-direction: column;
        justify-content: space-between;
        width: 100%;
        height: 100%;
        margin-top: 12px;
    }

    .add-input-container {
        display: flex;
        align-items: center;
        gap: 4px;
        width: unset !important;
    }

    .add-input-container > * {
        cursor: pointer;
    }

    .add-input-disabled > * {
        cursor: default !important;
        color: #777 !important;
    }

    #creation-row {
        justify-content: flex-start !important;
    }

    #create-button-wrapper {
        margin-left: 15px;
        height: 40px;
    }

    @media only screen and (max-width: 950px) {
        .manage {
            flex-direction: column;
        }

        .row {
            flex-direction: column;
        }

        #create-button-wrapper {
            margin-left: unset;
        }

        .form-select-row {
            max-height: unset;
            gap: 8px;
        }

        .multiselect-super {
            width: 100%;
        }

        .form-name-edit-wrapper {
            gap: unset;
        }

        .form-name-save-wrapper {
            width: 100%;
        }
    }
</style>
