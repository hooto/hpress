<div class="hp-comment-embed hp-block-gap-column">
  <div class="header"
    id="hp-comment-embed-list-header"
    style="display:{{if len .list.Items}}block{{else}}none{{end}}"
  >
    <nav class="nav-primary">
      <ul>
        <li>
          <span>Comments</span>
        </li>
      </ul>
    </nav>
  </div>

  <div id="hp-comment-embed-list" class="list">
    {{range $v := .list.Items}}
    <div class="entry">
      <div class="avatar">
        <img src='{{HttpSrvBasePath "hp/+/comment/~/img/user-default.png"}}' />
      </div>

      <div class="body">
        <div class="info">
          <strong>{{FieldSubString $v.Fields "author" 50}}</strong>
          <small>@{{UnixtimeFormat $v.Created "Y-m-d H:i"}}</small>
        </div>
        <p>{{FieldSubHtml $v.Fields "content" 2000}}</p>
      </div>
    </div>
    {{end}}
  </div>

  <div class="header">
    <nav class="nav-primary">
      <ul>
        <li>
          <span>New Comment</span>
        </li>
      </ul>
    </nav>
  </div>

  <div class="list">
    <div class="entry">
      <div class="avatar d-none d-lg-block">
        <img src='{{HttpSrvBasePath "hp/+/comment/~/img/user-default.png"}}' />
      </div>

      <div id="hp-comment-embed-new-form-ctrl" class="body">
        <div class="mb-2">
          <div class="info form-label"><strong>Guest</strong></div>
          <div>
            <input
              type="text"
              class="input form-control"
              name="author"
              placeholder="Leave a comment ..."
              onclick="hpComment.EmbedFormActive()"
            />
          </div>
        </div>
      </div>

      <div id="hp-comment-embed-new-form" class="body new mb-3" style="display: none">
        <input type="hidden" name="refer_id" value="{{.new_form_refer_id}}" />
        <input type="hidden" name="refer_modname" value="{{.new_form_refer_modname}}" />
        <input type="hidden" name="refer_datax_table" value="{{.new_form_refer_datax_table}}" />
        <input type="hidden" name="captcha_token" value="" />

        <div class="_form-group mb-2">
          <label class="form-label">Your name</label>
          <input
            type="text"
            class="input form-control"
            name="author"
            value="{{.new_form_author}}"
          />
        </div>

        <div class="_form-group mb-2">
          <label class="form-label">Content</label>
          <textarea class="textarea form-control" rows="3" name="content"></textarea>
        </div>

        <div class="_form-group mb-2">
          <label class="form-label">Verification</label>
          <div>
            <table width="100%">
              <tr>
                <td width="50%" valign="top">
                  <input type="text" class="input form-control" name="captcha_word" value="" />
                  <span class="form-text">Type the characters you see in the right picture</span>
                </td>
                <td style="width: 10px"></td>
                <td style="background-color: #dce6ff">
                  <img id="hp-comment-captcha-url" src="" />
                </td>
              </tr>
            </table>
          </div>
        </div>

        <div class="_form-group mb-2">
          <div id="hp-comment-embed-new-form-alert"></div>

          <div id="hp-comment-embed-new-form-footer">
            <button class="button is-dark btn btn-dark" onclick="hpComment.EmbedCommit()">
              Commit
            </button>
          </div>

          <div id="hp-comment-embed-new-form-footer-alert"></div>
        </div>
      </div>
    </div>
  </div>
</div>

<script id="hp-comment-embed-tpl" type="text/html">
  <div class="entry" id="entry-{[=it.meta.id]}">
    <div class="avatar">
      <img src="{[=hp.HttpSrvBasePath('+/comment/~/img/user-default.png')]}" />
    </div>
    <div class="body">
      <div class="info">
        <strong>{[=it.author]}</strong>
        <small>@{[=it.meta.created]}</small>
      </div>
      <p>{[=it.content]}</p>
    </div>
  </div>
</script>
