{#if $loadingScreen}
  <div id="loading-container" out:fade>
    <div class="loader">Wird Geladen...</div>
  </div>
{/if}

<script>
    import {loadingScreen} from '../js/stores'
    import { fade } from 'svelte/transition';
</script>

<style>
    #loading-container {
        position: fixed;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        background-color: #121212;
        z-index: 1001;

        display: flex;
        align-items: center;
    }

    .loader,
    .loader:before,
    .loader:after {
        background: #ffffff;
        -webkit-animation: load1 1s infinite ease-in-out;
        animation: load1 1s infinite ease-in-out;
        width: 1em;
        height: 4em;
    }
    .loader {
        color: #ffffff;
        text-indent: -9999em;
        margin: 88px auto;
        position: relative;
        font-size: 11px;
        -webkit-transform: translateZ(0);
        -ms-transform: translateZ(0);
        transform: translateZ(0);
        -webkit-animation-delay: -0.16s;
        animation-delay: -0.16s;
    }
    .loader:before,
    .loader:after {
        position: absolute;
        top: 0;
        content: '';
    }
    .loader:before {
        left: -1.5em;
        -webkit-animation-delay: -0.32s;
        animation-delay: -0.32s;
    }
    .loader:after {
        left: 1.5em;
    }
    @-webkit-keyframes load1 {
        0%,
        80%,
        100% {
            box-shadow: 0 0;
            height: 4em;
        }
        40% {
            box-shadow: 0 -2em;
            height: 5em;
        }
    }
    @keyframes load1 {
        0%,
        80%,
        100% {
            box-shadow: 0 0;
            height: 4em;
        }
        40% {
            box-shadow: 0 -2em;
            height: 5em;
        }
    }
</style>
