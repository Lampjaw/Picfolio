$(document).ready(function() {
    if ($('.image-container').length) {
        initPhotoGrid();

        var hashData = photoswipeParseHash();
        if(hashData.pid && hashData.gid) {
            openPhotoSwipe(hashData.pid, true);
        }

        window.onresize = function(e){
            initPhotoGrid();
        };
    }

    $('#album-editor-title').blur(function(event) {
        var modifiedValue = event.target.value;

        if (album.title === modifiedValue || (isEmpty(album.title) && isEmpty(modifiedValue))) {
            return;
        } else if (album.title && isEmpty(modifiedValue)) {
            event.target.value = album.title;
            return;
        } else {
            album.title = modifiedValue;
        }

        $.post('/album/' + album.id, album);
    });

    $('#album-editor-description').blur(function(event) {
        var modifiedValue = event.target.value;

        if (album.description === modifiedValue || (isEmpty(album.description) && isEmpty(modifiedValue))) {
            return;
        } else if (album.description && isEmpty(modifiedValue)) {
            album.description = null;
        } else {
            album.description = modifiedValue;
        }

        $.post('/album/' + album.id, album);
    });

    $('.image-editor-description').blur(function(event) {
        var photoId = event.target.closest('.image-editor').getAttribute('data-id');
        var modifiedValue = event.target.value;

        var currentPhoto = albumPhotos.find(a => a.id == photoId);

        if (currentPhoto.description === modifiedValue || (isEmpty(currentPhoto.description) && isEmpty(modifiedValue))) {
            return;
        } else if (currentPhoto.description && isEmpty(modifiedValue)) {
            currentPhoto.description = null;
        } else {
            currentPhoto.description = modifiedValue;
        }

        $.post('/image/' + photoId, currentPhoto);
    });

    $('.delete-album-confirm').click(function(event) {
        var albumId = event.target.getAttribute('data-id');
        $.ajax({
            url: '/album/' + albumId,
            type: 'DELETE',
            success: function() {
                window.location.href = "/admin";
            }
        });
    });

    $('.image-editor-delete-button').click(function() {
        var photoId = event.target.closest('.image-editor').getAttribute('data-id');
        $('#deleteImageModal').attr('data-id', photoId);

        $('#deleteImageModal').modal();
    });

    $('.delete-image-confirm').click(function(event) {
        var photoId = event.target.closest('#deleteImageModal').getAttribute('data-id');
        $(event.target.closest('#deleteImageModal')).removeAttr('data-id');

        $.ajax({
            url: '/image/' + photoId,
            type: 'DELETE',
            success: function() {
                location.reload();
            }
        });
    });

    $('.image-editor-cover-photo-button').click(function(event) {
        var photoId = event.target.closest('.image-editor').getAttribute('data-id');

        var oldCoverPhotoId = album.coverPhotoId;
        
        if (album.coverPhotoId === photoId) {
            return;
        } else {
            album.coverPhotoId = photoId;
        }

        $.post('/album/' + album.id, album, function() {
            $('.image-editor[data-id="' +  oldCoverPhotoId + '"] .image-editor-cover-photo-button').removeAttr("disabled");
            $('.image-editor[data-id="' +  album.coverPhotoId + '"] .image-editor-cover-photo-button').attr("disabled", true);
        });
    });

    $('.image-editor-rotate-button').click(function(event) {
        var photoId = event.target.closest('.image-editor').getAttribute('data-id');

        $.post('/image/' + photoId + '/rotate', function() {
            var imgElem = event.target.closest('.image-editor').querySelector('.image-editor-thumbnail');
            var imgSrc = imgElem.getAttribute('src');
            var d = new Date();
            $(imgElem).attr('src', imgSrc + '?' + d.getTime())
        });
    });
});

var isEmpty = function (str) {
    return (!str || 0 === str.length);
}

var initPhotoGrid = function() {
    var gridItems = [];
    
    if (gridMenu !== undefined) {
        gridItems.push(...gridMenu);
    }

    if (photos !== undefined) {
        gridItems.push(...photos);
    }

    $('.image-container').empty().justifiedImages({
        images : gridItems || [],
        rowHeight: 250,
        maxRowHeight: 500,
        margin: 2,
        thumbnailPath: function(photo) {
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
        template: function(photo) {
            if (photo.uploadButton) {
                return '<div class="photo-container menu-item" style="height:' + photo.displayHeight + 'px;margin-right:' + photo.marginRight + 'px;" data-toggle="modal" data-target="' + photo.menuTarget + '">' +
                    '<div class="image-thumb" style="width:' + photo.displayWidth + 'px;height:' + photo.displayHeight + 'px;" >' +
                    '<div class="menu-item-content">' +
                    '<div class="menu-icon">+</div>' +
                    '<div class="menu-text">' + photo.menuText + '</div>' +
                    '</div>' +
                    '</div>' +
                    '</div>';
            }

            return '<div class="photo-container swipeclick" style="height:' + photo.displayHeight + 'px;margin-right:' + photo.marginRight + 'px;" data-pid="' + photo.pid + '" >' +
                '<img class="image-thumb" src="' + photo.src + '" style="width:' + photo.displayWidth + 'px;height:' + photo.displayHeight + 'px;" >' +
                '</div>';
        }
    });

    $('.swipeclick').on("click", function(event) {
        var photoId = event.target.parentNode.getAttribute('data-pid');
        openPhotoSwipe(photoId);
    });
};

var openPhotoSwipe = function (photoId, disableAnimation) {
    var initPhotoIndex = photos.findIndex(function(p) {
        return p.pid === photoId;
    });

    var pswpElement = document.querySelectorAll('.pswp')[0];

    var options = {
        index: initPhotoIndex,
        galleryPIDs: true,
        getThumbBoundsFn: function(index) {
            var thumbnail = document.querySelector('.photo-container[data-pid="' + photos[index].pid + '"]');

            var pageYScroll = window.pageYOffset || document.documentElement.scrollTop;
            var rect = thumbnail.getBoundingClientRect(); 

            return {x:rect.left, y:rect.top + pageYScroll, w:rect.width};
        }
    };

    if(disableAnimation) {
        options.showAnimationDuration = 0;
    }

    var gallery = new PhotoSwipe(pswpElement, PhotoSwipeUI_Default, photos, options);
    gallery.init();
};

var photoswipeParseHash = function() {
    var hash = window.location.hash.substring(1),
    params = {};

    if(hash.length < 5) {
        return params;
    }

    var vars = hash.split('&');
    for (var i = 0; i < vars.length; i++) {
        if(!vars[i]) {
            continue;
        }
        var pair = vars[i].split('=');  
        if(pair.length < 2) {
            continue;
        }           
        params[pair[0]] = pair[1];
    }

    if(params.gid) {
        params.gid = parseInt(params.gid, 10);
    }

    return params;
};