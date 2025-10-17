<div id="hpm-specset" class="hpm-modal-formset">
  <div id="hpm-specset-alert"></div>

  <div class="d-flex flex-sm-row">
    <label class="form-label">Name</label>
    <div class="flex-fill">
      <input
        type="text"
        class="form-control form-control-sm"
        name="name"
        placeholder="Module Name"
        value="{[=it.meta.name]}"
        {[?
        it.meta.name]}
        readonly{[?]}
      />
    </div>
  </div>

  <div class="d-flex flex-sm-row">
    <label class="form-label">Service Path</label>
    <div class="flex-fill">
      <input
        type="text"
        class="form-control form-control-sm"
        name="srvname"
        placeholder="URL Prefix Name of Http Service"
        value="{[=it.srvname]}"
      />
      <div class="form-text">URL Prefix Name of Http Service</div>
    </div>
  </div>

  <div class="d-flex flex-sm-row">
    <label class="form-label">Title</label>
    <div class="flex-fill">
      <input
        type="text"
        class="form-control form-control-sm"
        name="title"
        placeholder="Title"
        value="{[=it.title]}"
      />
    </div>
  </div>

  {[if (it.meta.name != "core/general") {]}
  <div class="d-flex flex-sm-row">
    <label class="form-label">Status</label>
    <div class="flex-fill">
      <select class="form-select form-select-sm" name="status">
        <option value="1" {[if (it.status) { ]}selected{[ } ]}>Enable</option>
        <option value="0" {[if (!it.status) { ]}selected{[ } ]}>Disable</option>
      </select>
    </div>
  </div>
  {[}]}

  <div class="d-flex flex-sm-row">
    <label class="form-label">Theme</label>
    <div class="flex-fill">
      <textarea class="form-control form-control-sm" name="theme_config" rows="8">
{[? it.theme_config]}{[=it.theme_config]}{[?]}</textarea
      >
    </div>
  </div>
</div>
