$(document).ready(function() {
    var $form = $("#document_upload");
    var $doc_form = $("#form_document");

    $form.find('#UploadDoc').change(function() {
        $form.submit();
    });
    console.log("hmm");
    jajaxForm($form, function(r, status, xhr, $form) {
        if (!is_success(r)) {
            form_response_error(r);
            return;
        }
        form_hide_error();
        $doc_form.find("textarea").val(r.content.document);
        flash_success("Document converted.");
    });
})

