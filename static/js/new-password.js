function jpost(url, data) {
    return $.post(url, data, function() {}, "json");
}

function is_resp_err(data) {
    return data.status == 'error' || data.status == 'fail';
}

$(document).ready(function() {
    $form = $("#new-password");
    $nextbut = $form.find('.next');
    $submit = $form.find('input[type=submit]');
    $error = $form.find('.error');
    $herror = $form.find('.herror');
    $resend = $form.find('.resend');

    userid = $form.find('input[name=UserId]').val();

    // Prevent user from submitting form unless there is text in the
    // password box.
    $password = $form.find('input[name=Password]');
    $form.submit(function(ev) {
        if (!$password || !$password.val() || $password.val().length == 0) {
            ev.preventDefault();
        }
    });

    $nextbut.click(function() {
        jpost("/newpassword-json", $form.serialize())
        .done(function(r) {
            if (is_resp_err(r)) {
                $herror.html(r.message);
                $error.hide();
                $herror.show();
            } else {
                $herror.html("");
                $herror.hide();
                $error.show();
            }
        });
    });

    $resend.find('a').click(function() {
        jpost("/newpassword-send", { "UserId": userid })
        .done(function(r) {
            if (is_resp_err(r)) {
                $herror.html(r.message);
                $herror.show();
            } else {
                $herror.hide();
                $error.show();
                $resend.find('span.success').show();
            }
        });
    });
});
