<div class="guild-badge" on:click={invite}>
  <div class="guild-icon-bg">
    <i class="fas fa-plus fa-2x guild-icon-fa"></i>
  </div>

  <div style="padding-left: 10px">
    <span class="guild-name">Zu deinem Server einladen</span>
  </div>
</div>

<script>
    function invite() {
        window.location.href = `https://discord.com/oauth2/authorize?client_id=1319784485335728199`;
    }
</script>
