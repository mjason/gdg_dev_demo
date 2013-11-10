$.getJSON("/api/images", function(res) {
  window.res = res
  var source   = $("#entry-template").html()
  var template = Handlebars.compile(source)
  var html = template(res)
  $("body").append(html)
})