<Dropdown {col1} {col2} {col3} {col4} bind:value {label} {disabled} >
  {#if withNull}
    <option value=null>
      {nullLabel}
    </option>
  {/if}
  {#each channels as channel}
    {#if channel.type === 0 || (allowAnnouncementChannel && channel.type === 5)}
      <option value="{channel.id}">
        #{channel.name}
      </option>
    {/if}
  {/each}
</Dropdown>

<script>
    import Dropdown from "./form/Dropdown.svelte"

    export let value;
    export let label;
    export let disabled = false;
    export let channels = [];
    export let withNull = false;
    export let nullLabel = "Deaktiviert";
    export let allowAnnouncementChannel = false;

    $: value, ensureStringified();

    export let col1 = false;
    export let col2 = false;
    export let col3 = false;
    export let col4 = false;

    function ensureStringified() {
      if (value === null) {
        value = "null";
      }
    }
</script>
