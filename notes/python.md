# Python Notes

For this project, I didn't bother with `virtualenv` or anything like that. I just used Python from the Raspberry Pi repos, and installed dependencies using `pip3`.

## Updating

Is this really how people do this?

    pip3 freeze > requirements.txt
    sed -i 's/==/>=/g' requirements.txt
    pip3 install -r requirements.txt --upgrade

Also needed to install these when I last updated everything:

    sudo apt install libglib2.0-dev
    sudo apt install libgirepository1.0-dev
