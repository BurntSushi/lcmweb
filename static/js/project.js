function ready_my_projects() {
    $('.project-list .manage-collaborators').each(function() {
        var $this = $(this);
        var $form = $this.find('form');
        var $manage = $this.find('a.manage-start');
        var $done = $this.find('a.manage-done');
        var $collabs = $this.find('.collaborator-list');
        var proj_name = $form.find('input[name=ProjectName]').val();

        $manage.click(function(ev) {
            ev.preventDefault();
            if ($form.is(':visible')) {
                return;
            }

            viewable_show($form, ev);
        });
        $done.click(function(ev) {
            ev.preventDefault();
            viewable_hide($form);
        });
        $form.find('input[type=checkbox]').change(function() {
            $form.submit();
        });
        jajaxForm($form, function(r, stat, xhr, $form) {
            if (!is_success(r)) {
                flash_response_error(r);
                return;
            }

            var url = '/bit/{0}/{1}/collaborators'.format(User.Id, proj_name);
            html_get(url).done(function(data, stat, xhr) {
                $collabs.html(data);
            });
        });
    });
}

$(document).ready(function() {
    $form = $("#add-project");
    $submit = $form.find('input[type=submit]');
    $display_name = $form.find('#DisplayName');

    ready_my_projects();

    jajaxForm($form, function (r, status, xhr, $form) {
        if (!is_success(r)) {
            flash_response_error(r);
            return;
        }
        flash_hide_error();

        $display_name.val('');
        $display_name.blur();

        flash_success('Project <strong>' + r.content + '</strong> added.');
        html_get('/bit/myprojects').done(function(data, stat, xhr) {
            $('#my-projects').html(data);
            viewable_hide($form);
            ready_my_projects();
        });
    });

    $('#show-add-project').click(function(ev) {
        ev.preventDefault();
        viewable_show($form, ev, function() { $display_name.focus(); });
    });
    $form.find('#add-project-cancel').click(function() {
        viewable_hide($form);
        $display_name.blur();
    });
});
