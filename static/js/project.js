$(document).ready(function() {
    $form = $("#add-project");
    $submit = $form.find('input[type=submit]');

    jajaxForm($form, function (r, status, xhr, $form) {
        if (!is_success(r)) {
            form_response_error(r);
            return;
        }
        form_hide_error();
        console.log(r);
    });
});
