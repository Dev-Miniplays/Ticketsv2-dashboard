<Radio label="Benennungsschema" col4=true>
  <div class="radio-row">
    <input class="radio-input" type=radio bind:group={value} value="id">
    <label class="radio-label">#ticket-1</label>
  </div>

  <div class="radio-row">
    <input class="radio-input" type=radio bind:group={value} value="username">
    <label class="radio-label">#ticket-Benutzername</label>
  </div>
</Radio>

<script>
    import Radio from "./form/Radio.svelte";

    export let value;
</script>
