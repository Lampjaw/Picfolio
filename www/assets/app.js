$(document).ready(function() {
    $('.image-container').empty().justifiedImages({
        images : photos,
        rowHeight: 250,
        maxRowHeight: 500,
        thumbnailPath: function(photo, width, height) {
            return photo.msrc;
        },
        getSize: function(photo) {
            var aspect = photo.w / photo.h;

            var height = 0;
            var width = 0;

            if (photo.w > photo.h) {
                width = Math.min(photo.w, 650);
                height = width / aspect;
            } else if (photo.h > photo.w) {
                height = Math.min(photo.h, 650);
                width = height * aspect;
            } else {
                height = Math.min(photo.h, 650);
                width = Math.min(photo.w, 650);
            }

            return {width: width, height: height};
        },
        margin: 2
    });

    const regex = /.*\/(.*).thumb.jpg$/;

    $('.image-thumb').on("click", function(event) {
        var url = event.target.src;
        var imageId = url.match(regex)[1]
        
        var pidx = findPhotoById(imageId);

        openPhotoSwipe(pidx, event.target, event.target.parentNode)

    });

    var findPhotoById = function(imageId) {
        return photos.findIndex(function(p) {
            return p.pid === imageId;
        });
    }

    var openPhotoSwipe = function(index, thumbnailElement, galleryElement) {
        var pswpElement = document.querySelectorAll('.pswp')[0];

        options = {
            index: index,
            galleryPIDs: true,

            getThumbBoundsFn: function(index) {
                    pageYScroll = window.pageYOffset || document.documentElement.scrollTop,
                    rect = thumbnailElement.getBoundingClientRect(); 

                return {x:rect.left, y:rect.top + pageYScroll, w:rect.width};
            }

        };

        var gallery = new PhotoSwipe( pswpElement, PhotoSwipeUI_Default, photos, options);
        gallery.init();
    };
});